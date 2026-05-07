# Post-triggered Workflows

# Scoring Formulas

All scores are computed in **log₂-space** and composed **additively**. This makes scores numerically safe to store as Redis ZSET scores for the lifetime of the system, with no overflow, underflow, or recomputation needed for time decay. The product intuition (freshness should beat a doubling of quality after one half-life) is preserved exactly via the algebraic identity `log₂(2q) = log₂(q) + 1`.

---

## 1. Wilson Lower Bound (quality signal)

```
n  = likes + dislikes
p̂  = likes / n
WilsonLowerBound = ( p̂ + z²/2n − z × sqrt((p̂(1−p̂) + z²/4n) / n) )
                   ─────────────────────────────────────────────────
                                    1 + z²/n
z = 1.96  (95% confidence interval)

// n=0 (no votes):       WilsonLowerBound = 0  → quality term contributes 0
// n=1, 1 like:          ≈ 0.21  (low confidence, penalised)
// n=100, 80 likes:      ≈ 0.71
// n=100, 100 likes:     ≈ 0.96
```

---

## 2. Freshness Boost (linear time, additive)

Time enters as a **linear additive boost**, not a multiplicative decay. Newer posts score higher because their time term is larger; older posts don't actively decay, they're just outpaced as new posts arrive with larger boosts.

```
freshnessBoost(createdAt) = (createdAt_unixSeconds − EPOCH) / (halfLifeHours × 3600)

// EPOCH:         fixed reference timestamp (e.g. service launch)
//                keeps scores small numbers for years
// halfLifeHours: 168  (1 week)
```

**What `halfLifeHours = 168` means in this formulation:**

A post that is **one week newer** outranks a post with **2× the quality**. A post two weeks newer outranks one with 4× the quality, three weeks → 8×, and so on.

| Age (relative) | freshnessBoost contribution |
| -------------- | --------------------------- |
| now            | B (baseline)                |
| 1 week older   | B − 1                       |
| 2 weeks older  | B − 2                       |
| 4 weeks older  | B − 4                       |
| 8 weeks older  | B − 8                       |
| 1 year older   | B − ~52                     |

The numbers stay small and well-behaved indefinitely — a post 100 years old contributes only `B − ~5200`, easily inside double-precision float range.

---

## 3. Base Post Score

Quality is wrapped in `log₂` so a doubling of effective quality contributes exactly `+1`, matching the per-half-life freshness contribution.

```
qualityComponent = log₂( WilsonLowerBound(likes, dislikes) × Q_SCALE + 1 )

// Q_SCALE: maps Wilson's [0, 1] range into a useful range
//          default: 1000
//          → Wilson 0.71 contributes log₂(711) ≈ 9.5
//          → Wilson 0.96 contributes log₂(961) ≈ 9.9
//          → Wilson 0    contributes log₂(1)   = 0  (no votes)

postScore = SEED_BONUS + qualityComponent + freshnessBoost(createdAt)

// SEED_BONUS: small constant added to every post (e.g. 0.5)
//             does not affect ordering; just keeps absolute scores readable
```

**No special case needed for new posts.** When `n=0`, Wilson is 0 and `qualityComponent` is 0, so the post rides on `freshnessBoost` alone — exactly the right behavior. Newer no-vote posts naturally outrank older no-vote posts because their freshness term is larger.

**Examples** at `halfLifeHours=168`, `Q_SCALE=1000`, `SEED_BONUS=0.5`, expressed as differences from a freshly-posted no-vote baseline (the absolute number depends on EPOCH):

| State             | Age     | postScore (relative) |
| ----------------- | ------- | -------------------- |
| New, 0 votes      | 0h      | baseline             |
| New, 0 votes      | 1 week  | baseline − 1         |
| 10 votes, 80% up  | 0h      | baseline + ~9.0      |
| 10 votes, 80% up  | 1 week  | baseline + ~8.0      |
| 100 votes, 80% up | 0h      | baseline + ~9.5      |
| 100 votes, 80% up | 1 week  | baseline + ~8.5      |
| 100 votes, 80% up | 1 month | baseline + ~5.2      |
| 100 votes, 80% up | 1 year  | baseline − ~42       |

A 100-vote 80%-up post is worth roughly 9.5 freshness-units, so it beats any new no-vote post for ~9.5 weeks before being overtaken.

---

## 4. Distance Multiplier (follower tree & similarity tree)

Both trees use the same diminishing attenuation as depth increases, expressed in log-space so it composes additively with `postScore`:

```
distanceAttenuation(depth) = −log₂(depth)

// depth=1 (direct follower / 1-hop similar) →  0.00     (no attenuation)
// depth=2                                    → −1.00     (equiv. ×0.5)
// depth=3                                    → ~−1.58    (equiv. ×0.33)
// depth=n                                    → −log₂(n)
```

Mathematically equivalent to the multiplicative `1/depth` form, but additive composition keeps everything in the same units and avoids exponential/log round-trips.

---

## 5. Fan-out Score (per user, per path)

The score written into each follower's timeline ZSET:

```
// via follower tree:
fanOutScore = postScore − log₂(d_follow)

// via similarity tree:
fanOutScore = postScore − log₂(d_sim)

// user qualifies via both paths → take the closer connection (smaller depth):
fanOutScore = postScore − log₂( min(d_follow, d_sim) )
```

Because `postScore` is computed once at write time and the distance term is fixed by graph topology, **the ZSET score is stable**. Pagination via score-based cursors works correctly, and old posts are pushed down naturally as new posts arrive.

---

## 6. Interaction Multiplier (New Interaction workflow only)

Interaction weights become additive log-bonuses, applied to `postScore` before the distance term.

```
interactionBonus:
  like     → log₂(1.0) =  0.00
  comment  → log₂(1.2) ≈  0.26
  share    → log₂(1.5) ≈  0.58
  view     → log₂(0.3) ≈ −1.74

effectivePostScore = postScore + interactionBonus(type)
fanOutScore        = effectivePostScore − log₂(depth)
```

Same product behavior as the multiplicative form; expressed in the same additive units as everything else.

---

## 7. Score Change Threshold

Before any fan-out propagation, check if the score moved enough to warrant the work.

```
scoreDelta = | effectivePostScore − cachedPostScore |

if scoreDelta < SCORE_CHANGE_DELTA → abort, no update needed

// SCORE_CHANGE_DELTA: default 0.07  (≈ log₂(1.05), i.e. 5% effective change)
// cachedPostScore: stored in Redis at key  post_score:{postID}  (48h TTL)
// update cachedPostScore after passing threshold
```

To tune the threshold in human-readable terms:

```
SCORE_CHANGE_DELTA = log₂(1 + percent_threshold / 100)

//  5% →  0.0704
// 10% →  0.1375
// 25% →  0.3219
```

---

## 8. Score Update Propagation (the "re-fan-out" workflow)

When a post's `effectivePostScore` changes by more than `SCORE_CHANGE_DELTA` — typically because of new likes, comments, shares, or other engagement events — the new score must be propagated into the timelines of users who have this post in their feed but **haven't seen it yet**. Users who have already seen the post should not have it re-injected; the viewed-set filter handles that, but we also avoid the wasted ZSET write by checking up-front.

### 8.1 Trigger

Any workflow that updates `effectivePostScore` (interaction events, recount jobs, manual moderator actions) calls into the propagation pipeline after passing the threshold check in Section 7.

```
on score change:
  newScore = compute effectivePostScore
  oldScore = GET post_score:{postID}
  if | newScore − oldScore | < SCORE_CHANGE_DELTA:
      return                                     # not worth propagating
  SET post_score:{postID} = newScore  (TTL 48h)
  enqueue propagation job: { postID, newScore, oldScore }
```

### 8.2 The reverse index: who has this post?

Propagation needs to answer "which user timelines contain `postID`?" This is the inverse of the fan-out write and requires a **post → users reverse index**, maintained at fan-out time.

```
// at fan-out time, in addition to writing each user's timeline:
SADD post_recipients:{postID} user_id_1 user_id_2 ...
EXPIRE post_recipients:{postID} <timeline_retention_seconds>

// the TTL matches the timeline trim window (e.g. 7 days for a weekly trim,
// or longer if your timelines retain longer)
```

### 8.4 The viewed check matters operationally

For a post with high reach (say, 10 million recipients), the viewed-check pre-filter is what keeps propagation cheap:

- Without the check: 10M `ZREM` + 10M `ZADD` writes per score update = 20M Redis ops.
- With the check: ~10M `SISMEMBER` reads (cheap, in-memory bitset lookups) + writes only for the unviewed fraction.

If 70% of recipients have already viewed the post (typical for content that's been live for a few hours), the check eliminates 70% of the write load. The reads are much cheaper than the writes, so the net cost drops by ~60-70%.

For very high-reach posts, batch the propagation across workers (shard by user ID modulo worker count) so a single popular post's update doesn't bottleneck on one queue consumer.

### 8.5 Coalescing rapid updates

A post going viral can fire dozens of threshold-crossing events per minute, and each triggers a propagation job. Without coalescing, you can stampede the same recipient set with overlapping updates.

Mitigation:

```
on score change:
  if | newScore − oldScore | < SCORE_CHANGE_DELTA: return
  SET post_score:{postID} = newScore  (TTL 48h)

  # Debounce: only enqueue if no propagation job is already scheduled
  # for this post within the last DEBOUNCE_WINDOW seconds
  if SET propagation_lock:{postID} = 1 NX EX DEBOUNCE_WINDOW:
      enqueue propagation job
  # else: the in-flight or recently-queued job will pick up the latest score
  #       when it runs (it reads post_score:{postID} fresh)
```

The propagation job re-reads `post_score:{postID}` when it actually runs, so a debounced run still propagates the latest score, not whatever score triggered it. Effectively, multiple rapid changes collapse into one propagation pass.

`DEBOUNCE_WINDOW` should match how often you're willing to update — 30-60 seconds is reasonable for a feed; more aggressive for trending dashboards, less for archival systems.

### 8.6 Bounding the work: skip propagation for low-reach updates

Not every score change deserves propagation across all recipients. Two cheap bounds:

**Skip if reach is small enough to amortize on read.** If `SCARD post_recipients:{postID}` is below some threshold (say, 1000), the per-user write cost is small enough that you can just do it. Above that, consider whether the score change is large enough to justify the rewrite — the threshold can scale with reach:

```
SCORE_CHANGE_DELTA(reach) = base_threshold × log₂(reach / 1000)
// reach 1k     → base threshold (0.07)
// reach 100k   → ~0.5  (5x more conservative)
// reach 10M    → ~0.9  (much more conservative)
```

For high-reach posts, only large score swings trigger re-propagation.

**Skip if the post is near the bottom of recipients' timelines anyway.** If a post's score is already below the typical trim threshold, updating it further doesn't change the user-visible feed. This requires knowing the trim threshold per timeline, which is more complex than it's worth — usually skipped.

### 8.7 Full lifecycle summary

```
WRITE (initial fan-out):
  compute postScore → for each recipient:
    ZADD timeline:user:{userID} <fanOutScore> <postID>
    SADD post_recipients:{postID} <userID>
  EXPIRE post_recipients:{postID} <retention>
  SET post_score:{postID} = postScore  (TTL 48h)

INTERACTION (e.g. new like):
  recompute effectivePostScore
  if | new − cached | < SCORE_CHANGE_DELTA: abort
  SET post_score:{postID} = new
  acquire debounce lock; enqueue propagation job

PROPAGATION JOB (async worker):
  re-read post_score:{postID}  (latest score, in case of coalescing)
  recipients = SMEMBERS post_recipients:{postID}
  for batch in chunked(recipients, 500):
    bulk-check SISMEMBER viewed:user:{userID} <postID>
    for each unviewed user:
      ZREM  timeline:user:{userID} <postID>
      ZADD  timeline:user:{userID} XX <newScore> <postID>

READ (feed request):
  ZREVRANGEBYSCORE timeline:user:{userID} (cursor −inf LIMIT 0 N (overfetched)
  filter via SMISMEMBER viewed:user:{userID}
  return page; SADD viewed:user:{userID} <returned IDs>
```

---

## 9. Configuration Defaults

| Parameter             | Default                       | Meaning                                                     |
| --------------------- | ----------------------------- | ----------------------------------------------------------- | --- |
| `EPOCH`               | service launch (Unix seconds) | Reference timestamp for `freshnessBoost`                    |
| `halfLifeHours`       | 168 (1 week)                  | Time for freshness to equal a doubling of quality           |
| `Q_SCALE`             | 1000                          | Maps Wilson [0,1] into a useful additive range              |
| `SEED_BONUS`          | 0.5                           | Constant added to every post for absolute-score readability |
| `SCORE_CHANGE_DELTA`  | 0.07                          | log₂(1.05) — minimum score change to trigger propagation    |
| `DEBOUNCE_WINDOW`     | 30s                           | Coalescing window for rapid updates                         |
| `BATCH_SIZE`          | 500                           | Recipients per propagation batch                            |
| `post_recipients TTL` | matches timeline retention    | Garbage collection of reverse index                         |
| `post_score TTL`      | 48h                           | Cache TTL for the threshold check                           |
| `viewed:user TTL`     | matches timeline retention    | Bounds memory; matches what's filterable                    | --- |

## New Post

### Process the post

1. Compress photos to a different resolution as per database design (mobile, web).
2. Vectorize the post contents.
3. Link the post to other similar posts in the graph DB.
4. Trigger Fan-out Feed Workflow.

### Fan-out Feed Workflow

1. Get follower tree (configurable max depth) for the post author as a flat array with depth per entry.
2. Get similarity tree (configurable max depth) for the post as a flat array with depth per entry.
3. For each user in the follower tree:
   - Compute `fanOutScore = postScore × distanceMultiplier(d_follow)`.
   - Add post to user's feed with this score. Skip if post already in feed.
4. For each user reachable via the similarity tree (followers of users who interacted with similar posts):
   - Compute `fanOutScore = postScore × distanceMultiplier(d_sim)`.
   - If user already received post via step 3, keep `max(existing score, new score)`. Otherwise add.
5. Add similar posts to the feed of all reached users plus the post author:
   - Score each similar post using its own `postScore × distanceMultiplier(sim_depth)`.
   - Skip if already in feed.
6. Save the set of user IDs that received the post (for future score updates).

---

## New Interaction

1. Compute `effectivePostScore = postScore × interactionMultiplier(interactionType)`.
2. Compare against cached post score:
   ```
   scoreDelta = |effectivePostScore - cachedPostScore|
   if scoreDelta < scoreChangeDelta → skip this stage and continue with fan-out
   else → Score propagation to existing fan-out set:
        - Load the saved user IDs from the previous fan-out (step 7 of prior interactions / step 6 of New Post).
        - For each user in that saved set, update their feed score for this post:
     newFanOutScore = effectivePostScore × distanceMultiplier(depth_at_fan_out_time)
     if newFanOutScore > existingFeedScore → ZADD feed:{userID} newFanOutScore postID
   ```
   `scoreChangeDelta` default: `0.05`. Update cached score.
3. Get follower tree (configurable max depth) for the interacting user as a flat array with depth per entry.
4. Get similarity tree (configurable max depth) for the interacted post as a flat array with depth per entry.
5. For each user in the follower tree:
   - Compute `fanOutScore = effectivePostScore × distanceMultiplier(d_follow)`.
   - Add post in user's feed. Skip if user is in the set of users that received the fan-out.
6. For each user reachable via the similarity tree:
   - Compute `fanOutScore = effectivePostScore × distanceMultiplier(d_sim)`.
   - Add using `max(existing score, new score)`. Skip if user is in the set of users that received the fan-out.
7. Save the set of user IDs that received the interaction fan-out.

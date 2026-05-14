# Feed Score Calculation and Calibration

## 1. Overview

The score function ranks posts in a user's feed. It is computed once at fan-out time, stored as a Redis ZSET score, and never recomputed for time decay — posts age out implicitly as newer posts are inserted with higher scores.

The design rests on one insight: composing a logarithmic quality term with a linear time term produces a score whose components share a common unit and combine additively without numerical hazards.

**Part I** specifies the Wilson lower bound (the quality signal `q`). **Part II** specifies the score function and derives the calibration math.

---

# Part I — The Quality Signal

## 2. Wilson Lower Bound

### 2.1 Definition

```
n  = likes + dislikes
p̂  = likes / n
z  = 1.96    (95% confidence interval)

WilsonLowerBound(likes, dislikes) =
    ( p̂ + z²/(2n) − z · √( (p̂(1 − p̂) + z²/(4n)) / n ) )
    ─────────────────────────────────────────────────────
                       1 + z²/n
```

We denote this value `q`. By construction, `q ∈ [0, 1]`.

### 2.2 What Wilson does

The naive estimate `p̂ = likes / n` treats `9/10`, `90/100`, and `900/1000` as identical (all 90%), ignoring our differing confidence in each. Wilson returns the **lower bound of a 95% confidence interval** around the true rate: "given this data, what is the worst the true upvote rate could plausibly be?"

This produces two effects:

- **Penalizes low sample sizes.** `1/1` gives `q ≈ 0.21` despite `p̂ = 1.0`; `100/100` gives `q ≈ 0.96`.
- **Shrinks estimates toward 0.5** with diminishing strength as `n` grows.

### 2.3 How inputs affect the output

**Effect of `n` (holding `p̂ = 0.80` fixed):** `q` approaches `p̂` from below as `n` grows.

| `n` | 5     | 10    | 25    | 100   | 1000  | 10000 | ∞     |
| --- | ----- | ----- | ----- | ----- | ----- | ----- | ----- |
| `q` | 0.376 | 0.490 | 0.614 | 0.709 | 0.774 | 0.792 | 0.800 |

A post with `8/10` (q ≈ 0.49) and a post with `80/100` (q ≈ 0.71) share the same `p̂` but rank very differently — the second has earned more confidence.

**Effect of `p̂` (holding `n = 100` fixed):** monotonic, compressed near the extremes.

| `p̂` | 0.10  | 0.30  | 0.50  | 0.70  | 0.80  | 0.90  | 0.99  | 1.00  |
| --- | ----- | ----- | ----- | ----- | ----- | ----- | ----- | ----- |
| `q` | 0.055 | 0.218 | 0.402 | 0.602 | 0.709 | 0.824 | 0.946 | 0.962 |

**Per-vote impact is asymmetric and depends on state:**

| State        | +1 like      | +1 dislike      |
| ------------ | ------------ | --------------- |
| `(10, 0)`    | Δq = +0.019  | Δq = **−0.135** |
| `(50, 50)`   | Δq = +0.007  | Δq = −0.007     |
| `(900, 100)` | Δq = +0.0001 | Δq = −0.0006    |

Three implications: (1) at low `n`, dislikes hurt much more than likes help — they introduce uncertainty as well as lowering `p̂`; (2) at balanced `p̂`, votes are roughly symmetric; (3) at high `n`, individual votes barely move `q`. **Product consequence:** a few early downvotes can crater a fresh post and effectively bury it.

### 2.4 Why comments are excluded from `q`

There is a defensible case for treating comments as positive signal — commenting takes more effort than liking. Many production systems do. We deliberately don't, for two reasons.

**Technical:** Wilson is a confidence bound on a binomial proportion, requiring `numerator ≤ denominator`. Folding comments into the numerator naively (e.g., `p̂ = (likes + 2·comments) / n`) violates this and breaks the formula — `p̂` can exceed 1, making `p̂(1−p̂)` negative.

A statistically valid alternative exists — treating comments as fractional positive votes:

```
positive = likes + α · comments      (α ∈ [0, 1])
n        = positive + dislikes
p̂        = positive / n
```

But this requires choosing `α` — the assumed positivity of an average comment.

**Product:** Without comment-text sentiment analysis, we cannot estimate `α`. It varies sharply by content type:

- Friendly content (recipes, hobby, creators): `α ≈ 0.7–0.9`
- General social: `α ≈ 0.3–0.5`
- Contested or political: `α ≈ 0.1` or lower
- Inflammatory: `α` could plausibly be negative

A single global `α` would systematically misrank entire categories — most dangerously, treating heated arguments on inflammatory posts as endorsement. This is a well-documented failure mode of comment-counting feeds.

**Decision:** comments are excluded from `q` until a sentiment classifier is available. We accept the lost signal on friendly content in exchange for not amplifying outrage content. When sentiment classification ships, comments can be folded in via:

```
positive = likes + α · positive_comments
negative = dislikes + α · negative_comments
n        = positive + negative
p̂        = positive / n
```

### 2.5 Summary

| Property                    | Behavior                                |
| --------------------------- | --------------------------------------- |
| Range                       | `q ∈ [0, 1]`                            |
| At `n = 0`                  | `q = 0`                                 |
| Effect of more votes        | `q → p̂` from below                      |
| Effect of higher `p̂`        | monotonic, compressed near extremes     |
| Sensitivity to single votes | high at low `n`, negligible at high `n` |
| Asymmetry                   | dislikes > likes, especially at low `n` |
| Comments                    | excluded pending sentiment analysis     |
| What it estimates           | lower 95% CI bound on true upvote rate  |

---

# Part II — The Score Function

## 3. Definition

For a post with quality signal `q ∈ [0, 1]` and creation time `t` (Unix seconds):

```
score(q, t) = SEED_BONUS + log₂( max(q · Q_SCALE, 1) ) + (t − EPOCH) / (H · 3600)
```

| Symbol       | Meaning                                    |
| ------------ | ------------------------------------------ |
| `q`          | Wilson lower bound (Part I)                |
| `t`          | Post creation time, Unix seconds           |
| `EPOCH`      | Reference timestamp (e.g., service launch) |
| `H`          | Half-life parameter, in hours              |
| `Q_SCALE`    | Quality scaling factor                     |
| `SEED_BONUS` | Constant offset, identical for every post  |

Decomposed:

```
score = SEED_BONUS + Q(q) + F(t)

Q(q) = log₂( max(q · Q_SCALE, 1) )    quality component
F(t) = (t − EPOCH) / (H · 3600)        freshness component
```

`SEED_BONUS` is a **global constant** added to every post equally. Because all posts receive the same value, it is irrelevant to any ranking comparison — `score(A) − score(B)` cancels it out. It exists only to keep absolute scores in a readable positive range, and we ignore it in all analysis below.

The two ranking signals are post quality (`Q(q)`) and freshness (`F(t)`). Author identity, badges, reputation, and any other per-author attributes do **not** affect the score directly. Badges, if present in the system, exist for UI/profile display only and have no ranking effect.

### 3.1 The defensive floor

The `max(q · Q_SCALE, 1)` clamp ensures `Q(q) ≥ 0` for all inputs, including:

- `q = 0` (no votes), where `log₂(0) = −∞` would otherwise be a poison value
- `q < 1/Q_SCALE` (e.g., `q = 0.0001` from a single early downvote), where `log₂` would otherwise contribute large negative values that drown out freshness

For `q ≥ 1/Q_SCALE` the clamp is inactive and the formula reduces to `Q(q) = log₂(q · Q_SCALE)`. The calibration analysis below uses the unclamped form and assumes posts are in this regime.

---

## 4. The Common Unit

`Q(q)` and `F(t)` share a dimensionless **score unit**.

**Freshness unit:** by construction, `F(t + H · 3600) − F(t) = 1`. _One unit = one half-life of elapsed time._

**Quality unit:** by `log₂(2x) = log₂(x) + 1`, `Q(2q) − Q(q) = 1`. _One unit = one doubling of `q`._

**Commensurability:** because both terms use the same unit, they sum meaningfully. A score difference of `Δ` units between two posts means "post A is `Δ` doublings of quality, or `Δ` half-lives of freshness, or any equivalent combination, ahead of post B." This property comes entirely from the choice of `log₂` for quality and a linear time scale.

---

## 5. The Half-Life Equivalence Theorem

**Claim:** a post one half-life newer outranks a post with twice the quality.

### 5.1 Statement

For posts A, B with quality `q_A, q_B ≥ 1/Q_SCALE` and times `t_A, t_B`, let `Δt = t_A − t_B` and `r = q_B / q_A`. Then:

```
score(A) − score(B) = Δt / (H · 3600) − log₂(r)
```

So `score(A) ≥ score(B)` iff `Δt / (H · 3600) ≥ log₂(r)`.

### 5.2 Proof

```
score(A) − score(B)
  = [SEED + Q(q_A) + F(t_A)] − [SEED + Q(q_B) + F(t_B)]
  = log₂(q_A · Q_SCALE) − log₂(q_B · Q_SCALE) + (t_A − t_B) / (H · 3600)
  = log₂(q_A / q_B) + Δt / (H · 3600)
  = Δt / (H · 3600) − log₂(r)              ∎
```

Note that `SEED_BONUS` cancels out because it appears identically in both scores.

### 5.3 Corollary (the half-life property)

Setting `r = 2`, `Δt = H · 3600`: `score(A) − score(B) = 1 − 1 = 0`. **Exact tie.**

### 5.4 Generalization

For any real `k`: `Δt = kH · 3600` and `r = 2^k` gives `score(A) − score(B) = k − k = 0`.

| Age advantage | Quality multiplier compensated |
| ------------- | ------------------------------ |
| `1H` newer    | 2× quality                     |
| `2H` newer    | 4× quality                     |
| `kH` newer    | `2^k` × quality                |

These are exact equalities for any `q` in the unclamped regime.

---

## 6. Range and Numerical Bounds

| Quantity                          | Range                               |
| --------------------------------- | ----------------------------------- |
| `Q(q)`, unclamped                 | `[0, log₂(Q_SCALE)]`                |
| `F(t)` over retention window `W`  | `[−W/H, 0]` relative to now         |
| `F(t)` over 100 years (`H = 168`) | grows by ~5217 — within float range |

The score is numerically stable for any practical system lifetime. No overflow, no precision collapse, no need for periodic rescoring.

---

## 7. The Calibration Theorem

### 7.1 Statement

Let `q_max = 1` and `q_min = 1/Q_SCALE`. The **quality lifetime** — the largest age advantage at which a high-quality old post still outranks a fresh low-quality post — is:

```
L = log₂(Q_SCALE) · H hours
```

### 7.2 Proof

Apply Theorem 5.1 with A fresh at `q_min` and B older by `D` hours at `q_max`:

```
score(A) − score(B) = D/H − log₂(Q_SCALE)
```

A overtakes B iff `D > log₂(Q_SCALE) · H`. So `L = log₂(Q_SCALE) · H`. ∎

### 7.3 Calibration formula

Inverse: given desired `L` and `H`,

```
Q_SCALE = 2^(L / H)
```

---

## 8. Calibration Procedure

### 8.1 Pick the quality lifetime `L` (the main product decision)

| `L`       | Approx     | Feed character                                |
| --------- | ---------- | --------------------------------------------- |
| 24–72     | 1–3 days   | Very fresh; news, breaking events, live feeds |
| 168–504   | 1–3 weeks  | Balanced; general social                      |
| 1000–2000 | 6–12 weeks | Quality-favoring; professional, longform      |
| 4000+     | 6 months+  | Archival; reference, evergreen                |

### 8.2 Pick the half-life `H` (the aging speed)

| `H` | Approx  | Aging                       |
| --- | ------- | --------------------------- |
| 1–6 | hours   | Very fast; live event feeds |
| 24  | 1 day   | Daily feed cycle            |
| 168 | 1 week  | Weekly feed cycle           |
| 720 | 1 month | Monthly cycle               |

Smaller `H` produces larger `Q_SCALE` for a given `L`, which can yield extreme quality-component values. Pick `H` to fit the product cadence.

### 8.3 Derive `Q_SCALE`

```
Q_SCALE = 2^(L / H)
```

### 8.4 Worked examples

| Feed type  | `L` (hr)     | `H` (hr) | `Q_SCALE`     |
| ---------- | ------------ | -------- | ------------- |
| Slow / pro | 2016 (12 wk) | 168      | `2^12 = 4096` |
| General    | 1680 (10 wk) | 168      | `2^10 = 1024` |
| News       | 48 (2 days)  | 6        | `2^8 = 256`   |
| Live event | 6 (6 hr)     | 1        | `2^6 = 64`    |

### 8.5 Sanity check

Validate by computing scenarios via Theorem 5.1. Example with `Q_SCALE = 1024`, `H = 168`:

A 2-week-old viral post (`q = 0.95`):

```
Q(0.95) = log₂(0.95 · 1024) ≈ 9.93
F-loss  = 2 (two half-lives old)
Net advantage over fresh q_min post: ~7.93 units
Crossover: a fresh post wins if Q(q) ≥ 7.93, i.e., q ≥ 2^7.93 / 1024 ≈ 0.24.
```

So this 2-week-old viral post beats any fresh post with Wilson `< 0.24`. If this matches product intent, calibration is correct.

---

## 9. Distance Multiplier (follower tree & similarity tree)

The score so far (`SEED_BONUS + Q(q) + F(t)`) is a **per-post** value — every recipient of a given post sees the same `postScore`. To rank a post differently for different recipients based on how they relate to the author, we add a **distance attenuation** term computed at fan-out time.

A post can reach a user via two graph paths:

- **Follower tree:** the user follows the author (depth 1), or follows someone who follows the author (depth 2), and so on.
- **Similarity tree:** the user has interacted with similar posts (depth 1), or with users who interact with similar posts (depth 2), and so on.

Both paths use the same diminishing attenuation as depth increases, expressed in log-space so it composes additively with `postScore`:

```
distanceAttenuation(depth) = −log₂(depth)

// depth=1 (direct follower / 1-hop similar) →   0.00    (no attenuation)
// depth=2                                    →  −1.00    (equivalent to ×0.5)
// depth=3                                    →  ~−1.58   (equivalent to ×0.33)
// depth=n                                    →  −log₂(n)
```

Mathematically equivalent to the multiplicative `1/depth` form, but expressed additively to keep everything in the same score units and avoid round-trips through exponential/log space.

### 9.1 Why log₂(depth) specifically

Depth attenuation has the same "doubling" semantics as the rest of the score function:

- depth doubling → score drops by 1 unit
- 1-hop direct → 2-hop friend-of-friend: −1 unit (same as one half-life of age, or halving of `q`)
- 2-hop → 4-hop: another −1 unit

A user at depth `d` sees a post as if it were `log₂(d)` half-lives older than a direct follower would. This makes the distance term commensurable with quality and freshness — a 2-hop recipient effectively perceives the post one half-life "older" than a direct follower does, which is the right product behavior for graph-distance attenuation.

### 9.2 Alternative attenuation curves

`−log₂(depth)` falls off slowly for large depths (depth 16 is only −4 units, depth 1024 is −10). If you want more aggressive falloff, scale the term:

```
distanceAttenuation(depth) = −α · log₂(depth)    α ≥ 1
```

`α = 1` is the default. `α = 2` makes each depth doubling cost 2 units (depth 16 = −8). `α = 0.5` makes attenuation gentler.

Or use a different functional form entirely (e.g., `−depth`, linear in depth), but the log form is the natural choice because it composes with the rest of the system without introducing a new unit.

---

## 10. Fan-out Score (per user, per path)

The score written into each recipient's timeline ZSET combines `postScore` with the recipient-specific distance term:

```
// recipient reached via follower tree:
fanOutScore = postScore − log₂(d_follow)

// recipient reached via similarity tree:
fanOutScore = postScore − log₂(d_sim)

// recipient qualifies via both paths → take the closer connection (smaller depth):
fanOutScore = postScore − log₂( min(d_follow, d_sim) )
```

where `postScore = SEED_BONUS + Q(q) + F(t)` is the per-post value computed in Section 3.

The `min(d_follow, d_sim)` rule reflects "use the strongest available connection" — if the user is a direct follower of the author _and_ has interacted with similar posts at depth 3, count them as a direct follower (depth 1, no attenuation), not as 3-hop similar (−1.58 units).

### 10.1 Stability of the stored ZSET score

Because `postScore` is computed once at write time and the distance term is fixed by graph topology at fan-out time, **the ZSET score is permanent for that (post, recipient) pair**. Once written, it never changes:

- `Q(q)` is the Wilson value at fan-out time — frozen.
- `F(t)` depends only on `createdAt` and constants — never changes.
- `d_follow` and `d_sim` are graph distances at fan-out time — frozen for this fan-out event.

This preserves the design properties from Section 12.3: pagination via score-based cursors works correctly without snapshotting, old posts are pushed down naturally as new posts arrive with higher scores, and no recomputation is ever needed for time decay.

### 10.2 What changes when the graph changes

If a user later follows the author directly (changing `d_follow` from, say, 3 to 1), or unfollows, or the similarity graph updates, the **already-stored `fanOutScore` does not change**. Future posts by that author will use the new graph distance; existing posts in the user's timeline keep their original score.

This is the right tradeoff for almost all cases — recomputing every recipient's score every time the graph changes would be prohibitively expensive, and the inconsistency window (old posts ranked by stale graph distance) is invisible to users in practice. If specific cases demand correction (e.g., new follow should retroactively boost a viral post in the new follower's timeline), use the score-update propagation workflow on a per-event basis.

### 10.3 Range and bounds of `fanOutScore`

| Quantity              | Range                                         |
| --------------------- | --------------------------------------------- |
| `postScore`           | `SEED_BONUS + [0, log₂(Q_SCALE)] + F(t)`      |
| Distance term         | `(−∞, 0]` — always non-positive               |
| Practical depth bound | typically capped at ~6–8 hops at fan-out time |
| Distance term, capped | `[−log₂(d_max), 0]`                           |

`d_max` is the maximum graph depth at which a post is fanned out — a system parameter that caps the distance attenuation and bounds the fan-out cost per post. With `d_max = 8`, the distance term ranges in `[−3, 0]`.

### 10.4 Composition with the half-life equivalence

The half-life theorem (Section 5) extends naturally. For two recipients A and B at distances `d_A`, `d_B`:

```
fanOutScore(A) − fanOutScore(B) = Δt/(H · 3600) − log₂(r_q) − log₂(d_A / d_B)
```

So the same trade-off math applies: a doubling in graph distance is one score unit, equivalent to a half-life of age or a doubling of quality. A direct follower seeing an old post can be tied by a friend-of-friend (depth 2) seeing a half-life-newer post of the same quality.

---

## 11. Properties of the Quality Compressor

### 11.1 Logarithmic compression

Equal _ratios_ of quality produce equal score differences:

| Wilson change | Score change |
| ------------- | ------------ |
| 0.001 → 0.002 | +1           |
| 0.01 → 0.02   | +1           |
| 0.1 → 0.2     | +1           |
| 0.5 → 1.0     | +1           |

This compresses the high end of the quality range and expands the low end. **Product implication:** the formula is more aggressive at suppressing low-quality content than at distinguishing among high-quality content — which matches how user attention actually works.

### 11.2 Sensitivity

```
dQ/dq = 1 / (q · ln 2)
```

Large for small `q`, small near `q = 1`. This **compounds with the Wilson asymmetry** (Section 2.3): low-`n` posts are sensitive to single dislikes, _and_ low-`q` posts are sensitive to Wilson changes. Early dislikes on fresh posts produce double-amplified score swings.

### 11.3 Choice of base

Using `log₂` makes one score unit = doubling of `q`, matching half-life intuition. Other bases shift the equivalence:

| Base    | One unit equals     |
| ------- | ------------------- |
| `ln`    | ~2.72× quality      |
| `log₁₀` | 10× quality         |
| `log₂`  | 2× quality (chosen) |

`log₂` is conventional because `2×` matches "half-life as time-to-double."

---

## 12. Summary

### 12.1 Wilson signal

| Property            | Behavior                                |
| ------------------- | --------------------------------------- |
| Range               | `[0, 1]`                                |
| At `n = 0`          | `q = 0`                                 |
| More votes          | `q → p̂` from below                      |
| Higher upvote ratio | `q` increases monotonically             |
| Sensitivity         | high at low `n`, negligible at high `n` |
| Asymmetry           | dislikes hurt more than likes help      |
| Comments            | excluded pending sentiment analysis     |

### 12.2 Score function

| Quantity                   | Formula                                                |
| -------------------------- | ------------------------------------------------------ |
| Quality component          | `Q(q) = log₂(max(q · Q_SCALE, 1))`                     |
| Freshness component        | `F(t) = (t − EPOCH) / (H · 3600)`                      |
| Per-post score             | `postScore = SEED_BONUS + Q(q) + F(t)`                 |
| Distance attenuation       | `−log₂(depth)` per recipient                           |
| Per-recipient stored score | `fanOutScore = postScore − log₂(min(d_follow, d_sim))` |
| Score difference           | `Δscore = Δt/(H · 3600) − log₂(r_q) − log₂(d_A / d_B)` |
| Quality range              | `[0, log₂(Q_SCALE)]`                                   |
| Distance range             | `[−log₂(d_max), 0]`                                    |
| Freshness range over `W`   | `[−W/H, 0]`                                            |
| Half-life equivalence      | `kH` newer ⟺ `2^k` × quality OR `2^k` × graph-distance |
| Quality lifetime           | `L = log₂(Q_SCALE) · H` hours                          |
| Calibration                | `Q_SCALE = 2^(L/H)`                                    |

### 12.3 Design properties

1. **Numerically stable** for any practical system lifetime.
2. **No recomputation** needed for time decay; aging is implicit.
3. **Half-life equivalence is exact** (algebraic identity) for `q ≥ 1/Q_SCALE`.
4. **Two-knob calibration:** product behavior controlled by `L` and `H`; `Q_SCALE` is derived.
5. **Pagination-safe:** stored `fanOutScore` is immutable, enabling score-based cursors without snapshotting.
6. **Compositional:** distance and interaction multipliers compose as additive log-space terms without breaking any property.
7. **Defensive:** the `max(·, 1)` clamp handles zero and sub-threshold quality without business-logic guarantees on `q`.
8. **Author-agnostic:** ranking depends only on post quality, freshness, and graph distance; author identity, badges, and reputation do not affect the score. Cosmetic per-author indicators (badges in UI) are independent of the ranking system.

# Feed

## List

- Check if feed exists.
- If not, response with empty and trigger background feed update.
- If exists
  - Fetch exact amount of feed requested filter if cursor is provided
  - Iterate on the fetched feed and collect the unseen feeds only
  - If the final list is less than the requested amount, Fetch 2x the previous amount fetched. Repeat until either the feed is complete or the requested amount is reached
  - Cleanup the seen feed in the background for the future reads
  - Return the list of unseen feeds and a cursor to the next page

## Mark As Seen

- Add the post to the seen DB table
- Add the post to the seen feed Redis set with TTL from the Config

## Fan-out Feed Workflow (triggered by new post)

- Get follower tree (until the post score deeps bellow the minimum feed score) for the post author as a flat array with depth per entry.
- Get similarity tree (until the post score deeps bellow the minimum feed score) for the post as a flat array with depth per entry.
- For each user in the follower tree:
  - Compute `fanOutScore = postScore × distanceMultiplier(d_follow)`.
  - Add post to user's feed with this score. Skip if post already in feed or if score is too low.
- For each user reachable via the similarity tree (followers of users who interacted with similar posts):
  - Compute `fanOutScore = postScore × distanceMultiplier(d_sim)`.
  - If user already received post via step 3, keep `max(existing score, new score)`. Otherwise add.
- Add similar posts to the feed of all reached users plus the post author:
  - Score each similar post using its own `postScore × distanceMultiplier(sim_depth) × distanceMultiplier(d_follow)`.
  - Skip if already in feed.
- Save the set of user IDs that received the post (for future score updates).

## Update Feed Workflow (triggered by new interaction)

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

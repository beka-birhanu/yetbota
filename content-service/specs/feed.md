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

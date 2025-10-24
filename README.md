# catseye
 *A Cat's Eye View of Expiring VODs on NHK World.*

## To-do
-[] PWA support

## What it does
The "Expiring Soon" function on the NHK World TV app was removed following the new API changes.
This service restores such functionality using the new API.

## What it is made with
- [Bootstrap 5](https://getbootstrap.com/) (Frontend)
- Modified [nhk_expire](https://gist.github.com/metalfoxdev/bd9528f054b3ec18d1a813ad3517588c) written in Go (Backend)
- `humanize-duration.js` by [Evan Hahn](https://github.com/EvanHahn), licensed under the Unlicense

## How it works
The old API, known as J-Stream, had an API function for retrieving programmes expiring within a certain date range.
This function was used by the TV app.
Unfortanately, the new API does not have this function which is why it was removed from the TV app.

An alternate method has been found which involves the following steps:
1. Get list of categories (https://api.nhkworld.jp/showsapi/v1/en/categories/)
2. Get each category's `video_episodes` section (e.g. Anime and Manga: https://api.nhkworld.jp/showsapi/v1/en/categories/31/video_episodes)
3. Read `expired_at` for each episode
4. Profit

It's way slower than the old method, but it works.

## API (sort of)
You could use this service as an API by accessing the programme list from this URL: `https://metalfoxdev.github.io/catseye/progs.json`

## Disclaimer
This service is not endorsed in any way by NHK and/or it's subsidiaries.

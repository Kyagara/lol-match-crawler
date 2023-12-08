# lol-match-crawler

This is a very simple crawler for League of Legends matches purely for testing [equinox](https://github.com/Kyagara/equinox).

## Usage

```bash
EQUINOX_KEY="RGAPI..." DATABASE_URL="postgres://postgres:user@localhost:5432/database" go run . # or go build . && EQUINOX_KEY="" DATABASE_URL="" ./lol-match-crawler
```

## How it works

It starts by fetching the Korean challenger league with `league-v4` for 5v5 Ranked Solo queue, then it does the following:

- Loops through the league entries and gets each summoner with the `summoner-v4` API to acquire their PUUID.

- Loops through the summoners and gets their match list with 5 recent ranked 5x5 solo matches with `match-v5`.

- Loops through the matches and gets the match details and the match timeline with `match-v5`.

# gomodoro

Simple pomodoro timer :tomato:

## Usage

### Setup

#### Change time

Gomodoro reads `~/.gomodoro.yml` as config.

| key | type | description |
| :-- | :-- | :-- |
| pomodoro | Number | How long basic pomodoro will last |
| break | Number | How long break pomodoro( starts with `-b`) will last |
| pre | String | Scripts to run with `bash -lc 'XXX'` on starting a pomodoro |
| post | String | Scripts to run with `bash -lc 'XXX'` on ending a pomodoro |

There is a sample yml, `_sample/.gomodoro.yml`.

#### tmux status bar

With tmux status bar, gomodoro will show how long your pomodoro will last.

```
# tmux.conf
set -g status-right "#(cat ~/.gomodoro.tmux)"
set -g status-interval 1
```

That is bacause gomodoro writes remaining time on `~/.gomodoro.tmux`.

### Start a pomodoro

```
$ gomodoro &
```

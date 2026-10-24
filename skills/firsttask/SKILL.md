---
name: firsttask
description: Manage FirstTask (task management app) tasks, projects and tags with the firsttask CLI. Use for requests like "add a task", "show today's tasks", "mark this task as done" or "update the FirstTask item about X".
---

# FirstTask CLI

The `firsttask` command (https://github.com/suer/first-task-cli) operates on FirstTask's Firestore data as the user's Google account.

## Prerequisites

- Check that the command exists with `command -v firsttask`. If it is missing, ask the user to run `mise use -g github:suer/first-task-cli`.
- If a command fails with an authentication error or the user is not logged in, `firsttask login` is required. It opens a browser, so ask the user to run `! firsttask login` themselves. The config is at `~/.config/firsttask/config.json`.

## Commands

```sh
firsttask task list [--done] [--tag NAME] [--project TITLE] [-q TEXT] [--due-before DATE]
firsttask task show ID
firsttask task add TITLE [--memo M] [--project P] [--tag T]... [--start DATE] [--due DATE]
firsttask task edit ID [--title T] [--memo M] [--project P] [--tag T]... [--start DATE] [--due DATE]
firsttask task done ID
firsttask task undone ID
firsttask task rm ID
firsttask project list|add|rm
firsttask tag list|add|rm
```

- Dates are `YYYY-MM-DD`, `today` or `tomorrow`.
- An ID can be abbreviated to any prefix that identifies it uniquely.
- Always pass `--json` when you need to read the result.

## Workflow

- Before adding or updating, if a project or tag name is ambiguous, check the real names with `project list` / `tag list`.
- When the user refers to a task by title (e.g. "mark X as done"), look for candidates with `task list --json -q "X"`. If several match, ask the user; never guess when running `done`.
- `task rm`, `project rm` and `tag rm` cannot be undone. Show the target (title and ID) and get the user's confirmation before running them.
- After running a command, report the result briefly (the title and due date of the added or updated task, etc.).

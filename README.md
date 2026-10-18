# first-task-cli

CLI for managing [FirstTask](https://github.com/suer/first-task) data on Firestore.
It logs in as your Google account and talks to Firestore under the app's security rules.

## Setup

1. In the Firebase project's Google Cloud console, create an OAuth client of type **Desktop app**.
2. Create `~/.config/firsttask/config.json`:

   ```json
   {
     "apiKey": "<Firebase Web API key>",
     "projectId": "<Firebase project id>",
     "clientId": "<Desktop OAuth client id>",
     "clientSecret": "<Desktop OAuth client secret>"
   }
   ```

   Each key can also be set with `FIRSTTASK_API_KEY`, `FIRSTTASK_PROJECT_ID`, `FIRSTTASK_CLIENT_ID` and `FIRSTTASK_CLIENT_SECRET`.
3. `go build -o firsttask . && ./firsttask login`

## Usage

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

IDs can be abbreviated to a unique prefix. Dates are `YYYY-MM-DD`, `today` or `tomorrow`.
Add `--json` for machine-readable output.

---
description: Migrates an old, single-file feature document to the new directory structure.
---

Your task is to migrate a legacy single-file feature Markdown document to the multi-file directory structure.

The user has provided a file path as an argument: `$ARGUMENTS`.

**Steps:**

1.  Run the migration helper script using the Bash tool, passing the provided file path as the argument:
    ```
    $AI_SESSION_HOME/scripts/migrate_feature_file.sh "$ARGUMENTS"
    ```

2.  Report the outcome to the user: confirm the new directory was created, the files were populated, and the original file was archived with a `.migrated` extension.

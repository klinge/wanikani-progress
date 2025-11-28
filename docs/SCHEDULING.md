# Scheduling Daily Syncs

The WaniKani API application provides a `/api/sync` endpoint that can be triggered manually or automatically via system cron.

## Using System Cron (Recommended)

### Setup

1. Ensure the application is running and accessible
2. Add a cron job to trigger the sync endpoint

### Example Cron Configuration

```bash
# Edit crontab
crontab -e

# Add this line to sync daily at 2:00 AM
0 2 * * * curl -X POST http://localhost:8080/api/sync

# Or with error logging
0 2 * * * curl -X POST http://localhost:8080/api/sync >> /var/log/wanikani-sync.log 2>&1
```

### Cron Expression Examples

```bash
# Every day at 2:00 AM
0 2 * * *

# Every day at midnight
0 0 * * *

# Every 6 hours
0 */6 * * *

# Every Monday at 3:00 AM
0 3 * * 1

# Twice daily (6 AM and 6 PM)
0 6,18 * * *
```

### Using systemd Timer (Alternative)

If you prefer systemd timers over cron:

1. Create a service file `/etc/systemd/system/wanikani-sync.service`:

```ini
[Unit]
Description=WaniKani Data Sync
After=network.target

[Service]
Type=oneshot
ExecStart=/usr/bin/curl -X POST http://localhost:8080/api/sync
User=your-user
```

2. Create a timer file `/etc/systemd/system/wanikani-sync.timer`:

```ini
[Unit]
Description=Daily WaniKani Data Sync
Requires=wanikani-sync.service

[Timer]
OnCalendar=daily
OnCalendar=02:00
Persistent=true

[Install]
WantedBy=timers.target
```

3. Enable and start the timer:

```bash
sudo systemctl daemon-reload
sudo systemctl enable wanikani-sync.timer
sudo systemctl start wanikani-sync.timer

# Check timer status
sudo systemctl list-timers wanikani-sync.timer
```

## Monitoring Sync Results

### Check Sync Status

```bash
curl http://localhost:8080/api/sync/status
```

Response:
```json
{
  "syncing": false
}
```

### View Application Logs

The application logs sync operations with results. Check your application logs to see:
- When syncs were triggered
- How many records were updated
- Any errors that occurred

Example log output:
```
2024-01-15 02:00:01 INFO Sync started
2024-01-15 02:00:15 INFO Subjects synced: 150 records updated
2024-01-15 02:00:30 INFO Assignments synced: 75 records updated
2024-01-15 02:00:45 INFO Reviews synced: 200 records updated
2024-01-15 02:01:00 INFO Statistics synced: 1 record updated
2024-01-15 02:01:00 INFO Sync completed successfully
```

## Preventing Concurrent Syncs

The application automatically prevents concurrent sync operations. If a sync is already in progress when another is triggered, the second request will receive a 409 Conflict response:

```json
{
  "error": {
    "code": "SYNC_IN_PROGRESS",
    "message": "A sync operation is already in progress"
  }
}
```

This means you don't need to worry about cron jobs overlapping if a sync takes longer than expected.

## Troubleshooting

### Sync Not Running

1. Check if cron service is running:
   ```bash
   sudo systemctl status cron  # Debian/Ubuntu
   sudo systemctl status crond # RHEL/CentOS
   ```

2. Check cron logs:
   ```bash
   grep CRON /var/log/syslog  # Debian/Ubuntu
   grep CRON /var/log/cron    # RHEL/CentOS
   ```

3. Verify the application is accessible:
   ```bash
   curl http://localhost:8080/api/sync/status
   ```

### Sync Failing

1. Check application logs for error messages
2. Verify WaniKani API token is valid
3. Check network connectivity to WaniKani API
4. Verify database is accessible and has sufficient space

### Testing the Sync

Manually trigger a sync to test:

```bash
curl -X POST http://localhost:8080/api/sync
```

Expected response:
```json
{
  "message": "Sync completed successfully",
  "results": [
    {
      "DataType": "subjects",
      "RecordsUpdated": 150,
      "Success": true,
      "Error": "",
      "Timestamp": "2024-01-15T02:00:00Z"
    },
    ...
  ]
}
```

## Alternative: Programmatic Scheduler

If you prefer to have the scheduler built into the application (as originally specified in the requirements), you can implement task 9 from the tasks.md file. This would use the `robfig/cron` library to schedule syncs internally.

**Advantages of programmatic scheduler:**
- Self-contained application
- No external cron configuration needed
- Easier to test scheduling logic

**Advantages of system cron (current approach):**
- Simpler application code
- Easier to modify schedule without redeploying
- Standard Unix approach
- Works even if application restarts

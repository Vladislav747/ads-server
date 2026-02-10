Иногда пишет not enough space надо просто перезагрузить под



для clickhouse

```sql
create schema rotator;
```

```sql
CREATE TABLE statistics (
    ts TIMESTAMP,
    country Nullable(FixedString(2)),
    os Nullable(String),
    browser Nullable(String),
    campaign_id UInt32 DEFAULT 0,
    requests Int64 DEFAULT 0,
    impressions Int64 DEFAULT 0
)
ENGINE = MergeTree
PRIMARY KEY (ts, campaign_id);
```
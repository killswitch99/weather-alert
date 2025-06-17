#!/bin/bash
# Set database URL
export DATABASE_URL="postgresql://workflow:workflow123@localhost:5876/workflow_engine"
# Run migrations
psql $DATABASE_URL -f migrations/000001_init_workflows.up.sql

# Initialize sample data
psql $DATABASE_URL -f scripts/init_db.sql 
#!/usr/bin/bash


docker run -d \
  --name rabbitmq \
  -p 5672:5672 \
  -p 15672:15672 \
  -e RABBITMQ_DEFAULT_USER=guest \
  -e RABBITMQ_DEFAULT_PASS=guest \
  rabbitmq:3-management



docker run --name auth_db \
  -e POSTGRES_USER=auth_user\
  -e POSTGRES_PASSWORD=auth_pass \
  -e POSTGRES_DB=auth_db \
  -p 5432:5432 \
  -d postgres:16

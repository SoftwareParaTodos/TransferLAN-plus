#!/usr/bin/env bash
curl http://localhost:47231/queue
curl -X POST http://localhost:47231/queue/add -H "Content-Type: application/json" -d '{"file_name":"video.mp4","target":"PC-Max","size_bytes":5000000000}'

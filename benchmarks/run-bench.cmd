@echo off
start "" /high /affinity 4 /b /wait bench.test.exe -test.run=^$ -test.bench=. -test.benchtime=300ms -test.count=10 > go-bench-2026-11-02.txt

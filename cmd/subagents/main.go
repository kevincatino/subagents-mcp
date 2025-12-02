package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"subagents-mcp/internal/agents"
	"subagents-mcp/internal/logging"
	"subagents-mcp/internal/mcp"
	"subagents-mcp/internal/runner"
	"subagents-mcp/internal/validate"
)

func main() {
	agentsDirFlag := flag.String("agents-dir", "", "absolute path to agents directory containing YAML persona files")
	runnerFlag := flag.String("runner", "", "preferred runner (codex|copilot|gemini); leave blank to auto-select")
	runnerConfigFlag := flag.String("runner-config", "", "path to runner config yaml (optional)")
	flag.Parse()

	logger, err := logging.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync() //nolint:errcheck

	agentsDir, err := validate.Dir(*agentsDirFlag)
	if err != nil {
		logger.Fatal("invalid agents-dir", zap.Error(err))
	}

	repo := agents.NewYAMLRepository(agentsDir)

	var runnerConfig runner.Config
	if *runnerConfigFlag != "" {
		cfg, err := runner.LoadConfig(*runnerConfigFlag)
		if err != nil {
			logger.Fatal("invalid runner config", zap.Error(err))
		}
		runnerConfig = cfg
	}

	selector, err := runner.NewSelector(logger, runnerConfig, *runnerFlag)
	if err != nil {
		logger.Fatal("failed to construct runner", zap.Error(err))
	}

	server := mcp.NewServer(logger, repo, selector)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := server.Serve(ctx, os.Stdin, os.Stdout); err != nil {
		logger.Fatal("server stopped", zap.Error(err))
	}
}

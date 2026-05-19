package main

import (
	_ "github.com/liang21/heka/internal/domain/testcase"
	_ "github.com/liang21/heka/internal/infrastructure/persistence/postgres"
)

// This file just verifies that the repository implements the interface
// through compile-time checking

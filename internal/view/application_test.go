// Copyright 2025 The Toodofun Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http:www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package view

import (
	"context"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
)

func TestApplication_Run(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatalf("failed to init simulation screen: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	errCh := make(chan error)
	app := CreateApplication(ctx)
	app.SetScreen(screen)

	go func() {
		if err := app.Run(); err != nil {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		app.Stop()
	case err := <-errCh:
		t.Errorf("application run error: %s", err)
		app.Stop()
	}
}

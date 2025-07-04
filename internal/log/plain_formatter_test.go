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

package log

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestPlainFormatter_Format(t *testing.T) {
	formatter := &PlainFormatter{}
	entry := &logrus.Entry{
		Message: "test message",
	}

	formatted, err := formatter.Format(entry)

	assert.NoError(t, err)
	assert.Equal(t, "test message\n", string(formatted))
}

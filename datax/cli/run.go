// Copyright 2024 Eryx <evorui at gmail dot com>, All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cli

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/hooto/htoml4g/htoml"
)

var (
	client *dataxClient
	cfile  string
	cfg    ConfigCommon
	err    error
)

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

func resetPrompt(l *readline.Instance) {
	l.SetPrompt("cli : ")
}

func Run() error {

	if err := cfgSetup(); err != nil {
		return err
	}

	if err := mgrSetup(); err != nil {
		return err
	}

	l, err := readline.NewEx(&readline.Config{
		AutoComplete:        nil, // completer,
		HistoryFile:         fmt.Sprintf("~/.%s_history", "datax"),
		InterruptPrompt:     "^C",
		EOFPrompt:           "exit",
		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})
	if err != nil {
		panic(err)
	}
	defer l.Close()

	for {
		resetPrompt(l)

		line, err := l.Readline()

		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		out, err := cliInvoke(strings.TrimSpace(line), l)

		if err != nil {
			fmt.Println("Error:", err)
		} else if out != "" {
			fmt.Println(out)
		}
	}

	return nil
}

func cfgSetup() error {

	cfgFiles := []string{}
	for _, arg := range os.Args {
		if strings.HasSuffix(arg, ".toml") {
			cfgFiles = append(cfgFiles, arg)
		}
	}
	if len(cfgFiles) == 0 {
		return fmt.Errorf("no config file found")
	}

	var err error
	for _, cfgFile := range cfgFiles {
		if err = htoml.DecodeFromFile(cfgFile, &cfg); err == nil {
			cfile = cfgFile
			break
		}
	}
	if err != nil {
		return err
	}
	if len(cfg.Services) == 0 {
		return fmt.Errorf("no service endpoint config in %s", cfile)
	}

	var active *ConfigService
	for _, service := range cfg.Services {
		if cfg.LastActive != "" && cfg.LastActive == service.Name {
			active = service
			break
		}
	}
	if active == nil {
		active = cfg.Services[0]
	}

	if _, _, err := net.SplitHostPort(active.Addr); err != nil {
		return err
	}

	if client, err = active.NewClient(); err != nil {
		return err
	}

	if cfg.LastActive != active.Name {
		cfg.LastActive = active.Name
		htoml.EncodeToFile(&cfg, cfile)
	}

	fmt.Printf("connect to %s\n", active.Addr)

	return nil
}

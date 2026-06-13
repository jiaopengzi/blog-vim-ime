//
// FilePath    : blog-vim-ime\internal\ime\mode_test.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : IME 动作规则模块的单元测试.
//

package ime

import "testing"

func TestDetermineAction(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		before     string
		after      string
		wantAction TransitionAction
		wantErr    bool
	}{
		// normal -> any mode: switch to chinese
		{
			name:       "normal to insert switches chinese",
			before:     ModeNormal,
			after:      ModeInsert,
			wantAction: ActionSwitchToChinese,
		},
		{
			name:       "normal to visual switches chinese",
			before:     ModeNormal,
			after:      ModeVisual,
			wantAction: ActionSwitchToChinese,
		},
		{
			name:       "normal to replace switches chinese",
			before:     ModeNormal,
			after:      ModeReplace,
			wantAction: ActionSwitchToChinese,
		},
		{
			name:       "normal to cmd switches chinese",
			before:     ModeNormal,
			after:      ModeCmd,
			wantAction: ActionSwitchToChinese,
		},
		// any mode -> normal: switch to english
		{
			name:       "insert to normal switches english",
			before:     ModeInsert,
			after:      ModeNormal,
			wantAction: ActionSwitchToEnglish,
		},
		{
			name:       "visual to normal switches english",
			before:     ModeVisual,
			after:      ModeNormal,
			wantAction: ActionSwitchToEnglish,
		},
		{
			name:       "replace to normal switches english",
			before:     ModeReplace,
			after:      ModeNormal,
			wantAction: ActionSwitchToEnglish,
		},
		{
			name:       "cmd to normal switches english",
			before:     ModeCmd,
			after:      ModeNormal,
			wantAction: ActionSwitchToEnglish,
		},
		// same mode: no operation
		{
			name:       "normal to normal does nothing",
			before:     ModeNormal,
			after:      ModeNormal,
			wantAction: ActionNone,
		},
		{
			name:       "insert to insert does nothing",
			before:     ModeInsert,
			after:      ModeInsert,
			wantAction: ActionNone,
		},
		{
			name:       "visual to visual does nothing",
			before:     ModeVisual,
			after:      ModeVisual,
			wantAction: ActionNone,
		},
		// non-normal to non-normal: switch to chinese
		{
			name:       "insert to visual switches chinese",
			before:     ModeInsert,
			after:      ModeVisual,
			wantAction: ActionSwitchToChinese,
		},
		{
			name:       "visual to replace switches chinese",
			before:     ModeVisual,
			after:      ModeReplace,
			wantAction: ActionSwitchToChinese,
		},
		// invalid modes
		{
			name:    "invalid before mode",
			before:  "unknown",
			after:   ModeNormal,
			wantErr: true,
		},
		{
			name:    "invalid after mode",
			before:  ModeNormal,
			after:   "unknown",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := DetermineAction(tc.before, tc.after)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tc.wantAction {
				t.Fatalf("expected action %d, got %d", tc.wantAction, got)
			}
		})
	}
}

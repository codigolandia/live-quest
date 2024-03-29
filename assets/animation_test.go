package assets

import (
	"fmt"
	"testing"
)

type loadAnimTestCase struct {
	path string

	animationCount int
	frameCount     map[string]int
	skinsCount     map[string]int
}

var loadAnimTestCases = []loadAnimTestCase{
	{
		path: "gopher",

		animationCount: 3,
		frameCount: map[string]int{
			"standing":      5,
			"walking_left":  8,
			"walking_right": 8,
		},
		skinsCount: map[string]int{
			"standing": 5,
		},
	},
}

func TestLoadAnimation(t *testing.T) {
	for tn, tc := range loadAnimTestCases {
		t.Run(fmt.Sprintf("#%d", tn), func(t *testing.T) {
			bundle, err := LoadAnimation(tc.path)
			t.Logf("LoadAnimation(%s) => %v, %v", tc.path, bundle, err)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if bundle == nil {
				t.Fatalf("bundle is nil")
			}

			if len(bundle) != tc.animationCount {
				t.Errorf("unexpected animation count: want: %v, got: %v",
					tc.animationCount, len(bundle))
			}
			// all map keys were loaded
			for anim, count := range tc.frameCount {
				a, ok := bundle[anim]
				if !ok {
					t.Errorf("missing animation: %v", anim)
				} else {
					if len(a.Frames) != count {
						t.Errorf("wrong number of frames: want: %v, got: %v",
							count, len(a.Frames))
					}
				}
			}
			// all map keys were loaded
			for anim, count := range tc.skinsCount {
				a, ok := bundle[anim]
				if !ok {
					t.Errorf("missing animation: %v", anim)
				} else {
					if len(a.Skins) != count {
						t.Errorf("wrong number of frames: want: %v, got: %v",
							count, len(a.Skins))
					}
				}
			}
			// extra map keys were loaded
			for anim := range bundle {
				if _, ok := tc.frameCount[anim]; !ok {
					t.Errorf("extra animation loaded: %v", anim)
				}
			}
		})
	}
}

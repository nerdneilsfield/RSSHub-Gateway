package lb

import (
	"time"

	"github.com/nerdneilsfield/RSSHub-Gateway/internal/upstream"
)

type Picker interface {
	Pick(path string, now time.Time, avoid map[*upstream.State]struct{}) *upstream.State
}

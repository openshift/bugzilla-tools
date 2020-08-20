package eventlogger

import (
	"fmt"

	"github.com/openshift/library-go/pkg/operator/events"
	"k8s.io/klog"
)

type Recorder struct {
	component string
}

var _ events.Recorder = &Recorder{}

func NewRecorder(component string) events.Recorder {
	return &Recorder{
		component: component,
	}
}

func (r *Recorder) Event(reason, message string) {
	msg := fmt.Sprintf("[*%s#%s*] %s", r.component, reason, message)
	klog.Info(msg)
}

func (r *Recorder) Eventf(reason, messageFmt string, args ...interface{}) {
	r.Event(reason, fmt.Sprintf(messageFmt, args...))
}

func (r *Recorder) Warning(reason, message string) {
	msg := fmt.Sprintf(":warning: [*%s#%s*] %s", r.component, reason, message)
	klog.Warning(msg)
}

func (r *Recorder) Warningf(reason, messageFmt string, args ...interface{}) {
	r.Warning(reason, fmt.Sprintf(messageFmt, args...))
}

func (r *Recorder) ForComponent(componentName string) events.Recorder {
	newRecorder := *r
	newRecorder.component = componentName
	return &newRecorder
}

func (r *Recorder) WithComponentSuffix(componentNameSuffix string) events.Recorder {
	return r.ForComponent(r.component + "_" + componentNameSuffix)
}

func (r *Recorder) ComponentName() string {
	return r.component
}

func (r *Recorder) Shutdown() {}

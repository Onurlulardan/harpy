package deadletterdelivery

func (runner *Runner) setActive(session Session) {
	runner.lifecycle.mutex.Lock()
	defer runner.lifecycle.mutex.Unlock()
	runner.lifecycle.active = session
}

func (runner *Runner) detachActive() Session {
	runner.lifecycle.mutex.Lock()
	defer runner.lifecycle.mutex.Unlock()
	active := runner.lifecycle.active
	runner.lifecycle.active = nil
	return active
}

func closeSession(session Session) (err error) {
	if session == nil {
		return nil
	}
	defer func() {
		if recover() != nil {
			err = ErrSessionPanicked
		}
	}()
	return session.Close()
}

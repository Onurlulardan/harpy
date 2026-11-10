package config

func (c Worker) NotificationsKEKKeyring() (NotificationsKEKKeyring, error) {
	return c.notificationsKeys, c.notificationsKeys.Validate()
}

func (c Worker) Documents() (DocumentsWorker, error) {
	return c.documents, c.documents.Validate()
}

func (c Worker) LongRunning() LongRunning {
	return c.longRunning
}

func (c Worker) DeploymentMode() (DeploymentMode, error) {
	return c.deploymentMode, c.deploymentMode.Validate()
}

func (c Worker) Authority() WorkerAuthority { return c.authority }

func (c Worker) RabbitMQ() RabbitMQ { return c.rabbitMQ }

func (c Worker) MemoryLimit() WorkerMemoryLimit { return c.memoryLimit }

func (c Worker) IAM() (IAMWorker, error) {
	if err := c.iam.Validate(); err != nil {
		return IAMWorker{}, err
	}
	return c.iam, nil
}

func (c Worker) RestQueryCursorKeyring() (CursorKeyring, error) {
	if err := c.restQueryCursorKeys.Validate(); err != nil {
		return CursorKeyring{}, err
	}
	return c.restQueryCursorKeys, nil
}

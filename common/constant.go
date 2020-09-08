package common

const (
	// JobSaveDir : job save etcd path
	JobSaveDir = "/cron/jobs/"

	// JobKillerDir : job killer etcd path
	JobKillerDir = "/cron/killer/"

	// JobLockDir : job lock etcd path
	JobLockDir = "/cron/lock/"

	// JobWorkerDir : job worker register etcd path
	JobWorkerDir = "/cron/workers/"

	// JobEventSave : job event status 1 => save
	JobEventSave = 1

	// JobEventDelete : job evnet status 2 => delete
	JobEventDelete = 2

	// JobEventKill : job event status 3 => kill
	JobEventKill = 3
)

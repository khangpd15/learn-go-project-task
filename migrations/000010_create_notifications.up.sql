CREATE TABLE IF NOT EXISTS notifications (
    id SERIAL PRIMARY KEY,

    sender_id INT NOT NULL,
    receiver_id INT NOT NULL,

    type VARCHAR(50) NOT NULL,

    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,

    task_id INT NULL,
    project_id INT NULL,

    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    read_at TIMESTAMP NULL,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_notifications_type
        CHECK (type IN (
            'TASK_CREATED',
            'TASK_ASSIGNED',
            'TASK_UNASSIGNED',
            'TASK_STATUS_UPDATED',
            'TASK_DELETED',
            'PROJECT_UPDATED'
        )),

    CONSTRAINT fk_notifications_sender
        FOREIGN KEY (sender_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_notifications_receiver
        FOREIGN KEY (receiver_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_notifications_task
        FOREIGN KEY (task_id)
        REFERENCES tasks(id)
        ON DELETE SET NULL,

    CONSTRAINT fk_notifications_project
        FOREIGN KEY (project_id)
        REFERENCES projects(id)
        ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_notifications_receiver_created_at
    ON notifications(receiver_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_notifications_receiver_is_read_created_at
    ON notifications(receiver_id, is_read, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_notifications_sender_id
    ON notifications(sender_id);

CREATE INDEX IF NOT EXISTS idx_notifications_type
    ON notifications(type);

CREATE INDEX IF NOT EXISTS idx_notifications_task_id
    ON notifications(task_id);

CREATE INDEX IF NOT EXISTS idx_notifications_project_id
    ON notifications(project_id);

CREATE INDEX IF NOT EXISTS idx_notifications_unread_by_receiver
    ON notifications(receiver_id, created_at DESC)
    WHERE is_read = FALSE;
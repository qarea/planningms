CREATE TABLE Planning (
  PRIMARY KEY (id),
  id	            BIGINT                 NOT NULL AUTO_INCREMENT,
  user_id		    BIGINT                 NOT NULL,
  status            ENUM("OPEN","CLOSED")  NOT NULL,
  project_id        INT                    NOT NULL,
  tracker_id        INT                    NOT NULL,
  issue_id          INT		               NOT NULL,
  issue_title       VARCHAR(255)           NOT NULL,
  issue_url         VARCHAR(255)           NOT NULL,
  activity_id       INT                    NOT NULL,
  spent_online      INT                    NOT NULL,
  spent_offline     INT                    NOT NULL,
  reported          INT                    NOT NULL,
  created_at        BIGINT                 NOT NULL
);

CREATE TABLE PlannedTime (
  PRIMARY KEY (id),
  id	            BIGINT        NOT NULL AUTO_INCREMENT,
  planning_id       BIGINT        NOT NULL,
  estimation        INT           NOT NULL,
  reason            VARCHAR(255)  NOT NULL,
  created_at        BIGINT        NOT NULL,
  FOREIGN KEY (planning_id) REFERENCES Planning(id)
);

CREATE TABLE SpentTimeHistory (
  PRIMARY KEY (planning_id, started_at, status),
  planning_id       BIGINT                   NOT NULL,
  spent             INT                      NOT NULL,
  started_at        BIGINT                   NOT NULL,
  ended_at          BIGINT                   NOT NULL,
  status            ENUM("ONLINE","OFFLINE") NOT NULL,
  FOREIGN KEY (planning_id) REFERENCES Planning(id)
);

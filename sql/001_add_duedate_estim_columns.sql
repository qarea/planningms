ALTER TABLE Planning
  ADD issue_estim INT NOT NULL DEFAULT 0;
  
ALTER TABLE Planning
  ADD issue_due_date BIGINT NOT NULL DEFAULT 0;

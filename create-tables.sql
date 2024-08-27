-- DROP TABLE IF EXISTS emqx_messages;
-- CREATE TABLE emqx_messages (
--   sensor_id INT AUTO_INCREMENT PRIMARY KEY,
--   topic_name VARCHAR(128) NOT NULL,
--   measurement VARCHAR(128) NOT NULL,
--   last_measured timestamp DEFAULT NOW()
-- );

-- DROP TABLE IF EXISTS Sensors;
-- CREATE TABLE Sensors (
--   sensorID INT,
--   sensorName VARCHAR(128) NOT NULL,
--   address VARCHAR(128),
--   PRIMARY KEY (sensorID)
-- );

DROP TABLE IF EXISTS Topics;
CREATE TABLE Topics (
  topicID INT,
  topicName VARCHAR(128) NOT NULL,
  PRIMARY KEY (topicID)
);

DROP TABLE IF EXISTS Logs;
CREATE TABLE Logs (
  logID INT,
  topicID INT,
  measurement VARCHAR(128) NOT NULL,
  measureTime timestamp DEFAULT NOW(),
  PRIMARY KEY (logID),
  FOREIGN KEY (topicID) REFERENCES Topics(topicID)
);
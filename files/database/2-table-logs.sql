CREATE TABLE IF NOT EXISTS `logs` (
  `logid` int(11) NOT NULL AUTO_INCREMENT,
  `logday` date NOT NULL,
  `duration` float NOT NULL,
  `message` text NOT NULL,
  `state` enum('none','ready','synced') NOT NULL,
  PRIMARY KEY (`logid`)
);

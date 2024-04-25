CREATE TABLE IF NOT EXISTS `tags` (
  `tagid` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(80) NOT NULL,
  `value` varchar(120) NOT NULL,
  PRIMARY KEY (`tagid`)
);


CREATE  TABLE IF NOT EXISTS `logtags` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `logid` int(11) NOT NULL,
  `tagid` int(11) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `logid` (`logid`),
  KEY `tagid` (`tagid`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

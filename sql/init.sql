CREATE DATABASE `cronjob` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
use `cronjob`;

CREATE TABLE `cronjob_log` (
  `id` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `job_name` varchar(100) NOT NULL DEFAULT "",
  `command` varchar(1000) NOT NULL DEFAULT "",
  `err` text,
  `output` text,
  `plan_time` bigint(20) NOT NULL DEFAULT 0,
  `schedule_time` bigint(20) NOT NULL DEFAULT 0,
  `start_time` bigint(20) NOT NULL DEFAULT 0,
  `end_time`  bigint(20) NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  INDEX `index_job_name` (`job_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

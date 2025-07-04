CREATE
    DATABASE IF NOT EXISTS go_socket_chat_room DEFAULT CHARACTER SET utf8mb4;

DROP TABLE IF EXISTS `users`;
CREATE TABLE `users`
(
    id       VARCHAR(36) PRIMARY KEY,
    username VARCHAR(64)  NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL
) ENGINE = InnoDB
  CHARACTER SET = utf8mb4
  COLLATE = utf8mb4_unicode_ci
  ROW_FORMAT = Dynamic;
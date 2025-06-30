CREATE
DATABASE IF NOT EXISTS go_socket_chat_room DEFAULT CHARACTER SET utf8mb4;

DROP TABLE IF EXISTS `users`;
CREATE TABLE `users`  (
                          `id` int(0) NOT NULL AUTO_INCREMENT,
                          `username` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
                          `password` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
                          PRIMARY KEY (`id`) USING BTREE,
                          UNIQUE INDEX `username`(`username`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci ROW_FORMAT = Dynamic;
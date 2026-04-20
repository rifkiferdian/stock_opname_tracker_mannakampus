-- phpMyAdmin SQL Dump
-- version 5.2.1
-- https://www.phpmyadmin.net/
--
-- Host: 127.0.0.1
-- Generation Time: Apr 06, 2026 at 09:40 AM
-- Server version: 10.4.32-MariaDB
-- PHP Version: 8.2.12

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- Database: `asetra_manna_kampus`
--

-- --------------------------------------------------------

--
-- Table structure for table `model_has_permissions`
--

CREATE TABLE `model_has_permissions` (
  `permission_id` bigint(20) UNSIGNED NOT NULL,
  `model_type` varchar(255) NOT NULL,
  `model_id` bigint(20) UNSIGNED NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `model_has_roles`
--

CREATE TABLE `model_has_roles` (
  `role_id` bigint(20) UNSIGNED NOT NULL,
  `model_type` varchar(255) NOT NULL,
  `model_id` bigint(20) UNSIGNED NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Dumping data for table `model_has_roles`
--

INSERT INTO `model_has_roles` (`role_id`, `model_type`, `model_id`) VALUES
(1, 'Models\\User', 1),
(4, 'Models\\User', 6);

-- --------------------------------------------------------

--
-- Table structure for table `permissions`
--

CREATE TABLE `permissions` (
  `id` bigint(20) UNSIGNED NOT NULL,
  `name` varchar(255) NOT NULL,
  `group` varchar(255) DEFAULT NULL,
  `guard_name` varchar(255) NOT NULL,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Dumping data for table `permissions`
--

INSERT INTO `permissions` (`id`, `name`, `group`, `guard_name`, `created_at`, `updated_at`) VALUES
(1, 'permission_management_access', 'permission', 'web', '2025-09-30 20:23:01', '2025-09-30 20:23:01'),
(2, 'permission_view', 'permission', 'web', '2025-09-30 20:23:01', '2025-09-30 20:23:01'),
(3, 'permission_assign', 'permission', 'web', '2025-09-30 20:23:01', '2025-09-30 20:23:01'),
(4, 'permission_revoke', 'permission', 'web', '2025-09-30 20:23:01', '2025-09-30 20:23:01'),
(5, 'role_management_access', 'role', 'web', '2025-09-30 20:23:01', '2025-09-30 20:23:01'),
(6, 'role_view', 'role', 'web', '2025-09-30 20:23:01', '2025-09-30 20:23:01'),
(7, 'role_create', 'role', 'web', '2025-09-30 20:23:01', '2025-09-30 20:23:01'),
(8, 'role_edit', 'role', 'web', '2025-09-30 20:23:01', '2025-09-30 20:23:01'),
(9, 'role_delete', 'role', 'web', '2025-09-30 20:23:01', '2025-09-30 20:23:01'),
(10, 'user_management_access', 'user', 'web', '2025-09-30 20:23:01', '2025-09-30 20:23:01'),
(11, 'user_view', 'user', 'web', '2025-09-30 20:23:01', '2025-09-30 20:23:01'),
(12, 'user_create', 'user', 'web', '2025-09-30 20:23:01', '2025-09-30 20:23:01'),
(13, 'user_edit', 'user', 'web', '2025-09-30 20:23:01', '2025-09-30 20:23:01'),
(14, 'user_delete', 'user', 'web', '2025-09-30 20:23:01', '2025-09-30 20:23:01'),
(15, 'system_settings_access', 'system_settings', 'web', '2025-09-30 20:23:01', '2025-09-30 20:23:01'),
(16, 'app_settings_manage', 'app_settings', 'web', '2025-09-30 20:23:01', '2025-09-30 20:23:01');

-- --------------------------------------------------------

--
-- Table structure for table `roles`
--

CREATE TABLE `roles` (
  `id` bigint(20) UNSIGNED NOT NULL,
  `name` varchar(255) NOT NULL,
  `guard_name` varchar(255) NOT NULL,
  `is_admin` tinyint(1) NOT NULL DEFAULT 0,
  `created_at` timestamp NULL DEFAULT current_timestamp(),
  `updated_at` timestamp NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Dumping data for table `roles`
--

INSERT INTO `roles` (`id`, `name`, `guard_name`, `is_admin`, `created_at`, `updated_at`) VALUES
(1, 'super-admin', 'web', 0, '2025-09-30 20:23:01', '2025-09-30 20:23:01'),
(2, 'admin', 'web', 0, '2025-09-30 20:23:01', '2025-09-30 20:23:01'),
(3, 'manager', 'web', 0, '2025-11-11 08:11:46', '2025-11-11 08:11:46'),
(4, 'staff-counter', 'web', 0, '2025-10-24 00:31:37', '2025-10-24 00:31:37');

-- --------------------------------------------------------

--
-- Table structure for table `role_has_permissions`
--

CREATE TABLE `role_has_permissions` (
  `permission_id` bigint(20) UNSIGNED NOT NULL,
  `role_id` bigint(20) UNSIGNED NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Dumping data for table `role_has_permissions`
--

INSERT INTO `role_has_permissions` (`permission_id`, `role_id`) VALUES
(1, 1),
(1, 3),
(2, 1),
(2, 3),
(3, 1),
(3, 3),
(4, 1),
(4, 3),
(5, 1),
(5, 3),
(6, 1),
(6, 3),
(7, 1),
(7, 3),
(8, 1),
(8, 3),
(9, 1),
(9, 3),
(10, 1),
(10, 3),
(11, 1),
(11, 3),
(12, 1),
(12, 3),
(13, 1),
(13, 3),
(14, 1),
(14, 3),
(15, 1),
(15, 3),
(15, 4),
(16, 1),
(16, 3);

-- --------------------------------------------------------

--
-- Table structure for table `stores`
--

CREATE TABLE `stores` (
  `store_id` int(11) NOT NULL,
  `store_code` varchar(255) NOT NULL,
  `store_name` varchar(255) NOT NULL,
  `store_address` text NOT NULL,
  `is_active` tinyint(1) NOT NULL DEFAULT 1,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `updated_at` datetime DEFAULT current_timestamp() ON UPDATE current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `stores`
--

INSERT INTO `stores` (`store_id`, `store_code`, `store_name`, `store_address`, `is_active`, `created_at`, `updated_at`) VALUES
(1, 'MK1', 'MK1 Babarsari', 'Babarsari', 1, '2025-12-18 21:01:02', '2025-12-19 11:01:02'),
(2, 'MK2', 'MK2 Simanjuntak', 'Simanjuntak', 1, '2025-12-18 21:01:02', '2025-12-19 11:01:02'),
(3, 'MK3', 'MK3 Supeno', 'Supeno', 1, '2025-12-18 21:01:02', '2025-12-19 11:01:02'),
(4, 'MK4', 'MK4 Palagan', 'Palagan', 1, '2025-12-18 21:01:02', '2025-12-19 11:01:02'),
(5, 'MK5', 'MK5 Godean', 'Godean', 1, '2025-12-18 21:01:02', '2025-12-19 11:01:02'),
(6, 'MK6', 'MK6 Imogiri', 'Imogiri', 1, '2025-12-18 21:01:02', '2025-12-19 11:01:02'),
(7, 'MK7', 'MK7 Keloran', 'Keloran', 1, '2025-12-18 21:01:02', '2025-12-19 11:01:02'),
(101, 'MKM1', 'MK Mini 1 Pelemsewu', 'Pelemsewu', 1, '2025-12-18 21:01:02', '2025-12-19 11:01:02'),
(102, 'MKM2', 'MK Mini 2 Diro', 'Diro', 1, '2025-12-18 21:01:02', '2025-12-19 11:01:02'),
(103, 'MKM3', 'MK Mini 3 Minomartani', 'Minomartani', 1, '2025-12-18 21:01:02', '2025-12-19 11:01:02');

-- --------------------------------------------------------

--
-- Table structure for table `users`
--

CREATE TABLE `users` (
  `id` int(11) NOT NULL,
  `nip` int(11) NOT NULL,
  `username` varchar(255) NOT NULL,
  `password` varchar(255) NOT NULL,
  `name` varchar(255) NOT NULL,
  `email` varchar(255) DEFAULT NULL,
  `status` enum('active','non_active') DEFAULT 'active',
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `updated_at` timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `users`
--

INSERT INTO `users` (`id`, `nip`, `username`, `password`, `name`, `email`, `status`, `created_at`, `updated_at`) VALUES
(1, 250192, 'admin', '$2a$10$d2fwsVDPcTGsI10DM67KSe6CFn7UyMHuHTGATyBKK770Dh2EZf/Qu', 'Admin Rifki', 'admin@mannakampus.com', 'active', '2025-11-25 07:42:56', '2026-01-03 02:46:21');

-- --------------------------------------------------------

--
-- Table structure for table `user_stores`
--

CREATE TABLE `user_stores` (
  `user_id` int(11) NOT NULL,
  `store_id` int(11) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `user_stores`
--

INSERT INTO `user_stores` (`user_id`, `store_id`) VALUES
(1, 1),
(1, 2),
(1, 3),
(1, 4),
(1, 5),
(1, 6),
(1, 7);

--
-- Indexes for dumped tables
--

--
-- Indexes for table `model_has_permissions`
--
ALTER TABLE `model_has_permissions`
  ADD PRIMARY KEY (`permission_id`,`model_id`,`model_type`),
  ADD KEY `model_has_permissions_model_id_model_type_index` (`model_id`,`model_type`);

--
-- Indexes for table `model_has_roles`
--
ALTER TABLE `model_has_roles`
  ADD PRIMARY KEY (`role_id`,`model_id`,`model_type`),
  ADD KEY `model_has_roles_model_id_model_type_index` (`model_id`,`model_type`);

--
-- Indexes for table `permissions`
--
ALTER TABLE `permissions`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `permissions_name_guard_name_unique` (`name`,`guard_name`);

--
-- Indexes for table `roles`
--
ALTER TABLE `roles`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `roles_name_guard_name_unique` (`name`,`guard_name`);

--
-- Indexes for table `role_has_permissions`
--
ALTER TABLE `role_has_permissions`
  ADD PRIMARY KEY (`permission_id`,`role_id`),
  ADD KEY `role_has_permissions_role_id_foreign` (`role_id`);

--
-- Indexes for table `stores`
--
ALTER TABLE `stores`
  ADD PRIMARY KEY (`store_id`),
  ADD KEY `store_id` (`store_id`);

--
-- Indexes for table `users`
--
ALTER TABLE `users`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `username` (`username`),
  ADD UNIQUE KEY `username_2` (`username`),
  ADD UNIQUE KEY `nip` (`nip`),
  ADD UNIQUE KEY `email` (`email`) USING BTREE,
  ADD KEY `nip_2` (`nip`);

--
-- Indexes for table `user_stores`
--
ALTER TABLE `user_stores`
  ADD PRIMARY KEY (`user_id`,`store_id`),
  ADD KEY `store_id` (`store_id`);

--
-- AUTO_INCREMENT for dumped tables
--

--
-- AUTO_INCREMENT for table `permissions`
--
ALTER TABLE `permissions`
  MODIFY `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=53;

--
-- AUTO_INCREMENT for table `roles`
--
ALTER TABLE `roles`
  MODIFY `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=9;

--
-- AUTO_INCREMENT for table `users`
--
ALTER TABLE `users`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=7;

--
-- Constraints for dumped tables
--

--
-- Constraints for table `model_has_permissions`
--
ALTER TABLE `model_has_permissions`
  ADD CONSTRAINT `model_has_permissions_ibfk_1` FOREIGN KEY (`permission_id`) REFERENCES `permissions` (`id`) ON DELETE CASCADE;

--
-- Constraints for table `model_has_roles`
--
ALTER TABLE `model_has_roles`
  ADD CONSTRAINT `model_has_roles_ibfk_1` FOREIGN KEY (`role_id`) REFERENCES `roles` (`id`) ON DELETE CASCADE;

--
-- Constraints for table `role_has_permissions`
--
ALTER TABLE `role_has_permissions`
  ADD CONSTRAINT `role_has_permissions_ibfk_1` FOREIGN KEY (`permission_id`) REFERENCES `permissions` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `role_has_permissions_ibfk_2` FOREIGN KEY (`role_id`) REFERENCES `roles` (`id`) ON DELETE CASCADE;

--
-- Constraints for table `user_stores`
--
ALTER TABLE `user_stores`
  ADD CONSTRAINT `user_stores_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  ADD CONSTRAINT `user_stores_ibfk_2` FOREIGN KEY (`store_id`) REFERENCES `stores` (`store_id`) ON DELETE CASCADE ON UPDATE CASCADE;
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;

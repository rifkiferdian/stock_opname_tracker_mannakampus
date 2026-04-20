-- phpMyAdmin SQL Dump
-- version 5.2.1
-- https://www.phpmyadmin.net/
--
-- Host: 127.0.0.1
-- Generation Time: Apr 20, 2026 at 04:09 AM
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
-- Database: `stock_opname_tracker`
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
-- Table structure for table `products`
--

CREATE TABLE `products` (
  `id` int(11) NOT NULL,
  `product_code` varchar(50) NOT NULL,
  `barcode` varchar(100) DEFAULT NULL,
  `product_name` varchar(200) NOT NULL,
  `category_id` int(11) DEFAULT NULL,
  `unit_id` int(11) DEFAULT NULL,
  `brand` varchar(100) DEFAULT NULL,
  `min_stock` decimal(18,2) NOT NULL DEFAULT 0.00,
  `max_stock` decimal(18,2) NOT NULL DEFAULT 0.00,
  `reorder_point` decimal(18,2) NOT NULL DEFAULT 0.00,
  `default_lead_time_days` int(11) NOT NULL DEFAULT 0,
  `pack_size` decimal(18,2) NOT NULL DEFAULT 1.00,
  `is_active` tinyint(1) NOT NULL DEFAULT 1,
  `created_by` int(11) DEFAULT NULL,
  `updated_by` int(11) DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `updated_at` timestamp NULL DEFAULT current_timestamp() ON UPDATE current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `products`
--

INSERT INTO `products` (`id`, `product_code`, `barcode`, `product_name`, `category_id`, `unit_id`, `brand`, `min_stock`, `max_stock`, `reorder_point`, `default_lead_time_days`, `pack_size`, `is_active`, `created_by`, `updated_by`, `created_at`, `updated_at`) VALUES
(1, 'BRG-0001', '899999000001', 'Indomie Goreng', 1, 1, 'Indomie', 20.00, 200.00, 30.00, 2, 1.00, 1, 1, NULL, '2026-04-17 06:09:43', '2026-04-17 06:09:43'),
(2, 'BRG-0002', '899999000002', 'Sabun Mandi ABC', 2, 1, 'ABC', 10.00, 100.00, 15.00, 3, 1.00, 1, 1, NULL, '2026-04-17 06:09:43', '2026-04-17 06:09:43'),
(3, 'BRG-0003', '899999000003', 'Tisu Wajah', 3, 1, 'Nice', 15.00, 120.00, 20.00, 2, 1.00, 1, 1, NULL, '2026-04-17 06:09:43', '2026-04-17 06:09:43');

-- --------------------------------------------------------

--
-- Table structure for table `product_categories`
--

CREATE TABLE `product_categories` (
  `id` int(11) NOT NULL,
  `category_code` varchar(50) NOT NULL,
  `category_name` varchar(150) NOT NULL,
  `description` text DEFAULT NULL,
  `is_active` tinyint(1) NOT NULL DEFAULT 1,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `updated_at` timestamp NULL DEFAULT current_timestamp() ON UPDATE current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `product_categories`
--

INSERT INTO `product_categories` (`id`, `category_code`, `category_name`, `description`, `is_active`, `created_at`, `updated_at`) VALUES
(1, 'FOOD', 'Food', 'Produk makanan', 1, '2026-04-17 04:16:10', '2026-04-17 04:16:10'),
(2, 'TOY', 'Toiletries', 'Produk toiletris', 1, '2026-04-17 04:16:10', '2026-04-17 04:16:10'),
(3, 'DEPT', 'Dept Store', 'Produk kebutuhan umum', 1, '2026-04-17 04:16:10', '2026-04-17 04:16:10');

-- --------------------------------------------------------

--
-- Table structure for table `product_suppliers`
--

CREATE TABLE `product_suppliers` (
  `id` int(11) NOT NULL,
  `product_id` int(11) NOT NULL,
  `supplier_id` int(11) NOT NULL,
  `is_primary` tinyint(1) NOT NULL DEFAULT 0,
  `priority_no` int(11) NOT NULL DEFAULT 1,
  `last_price` decimal(18,2) NOT NULL DEFAULT 0.00,
  `moq` decimal(18,2) NOT NULL DEFAULT 0.00,
  `lead_time_days` int(11) NOT NULL DEFAULT 0,
  `pack_size` decimal(18,2) NOT NULL DEFAULT 1.00,
  `is_active` tinyint(1) NOT NULL DEFAULT 1,
  `notes` varchar(255) DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `updated_at` timestamp NULL DEFAULT current_timestamp() ON UPDATE current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `product_suppliers`
--

INSERT INTO `product_suppliers` (`id`, `product_id`, `supplier_id`, `is_primary`, `priority_no`, `last_price`, `moq`, `lead_time_days`, `pack_size`, `is_active`, `notes`, `created_at`, `updated_at`) VALUES
(1, 1, 1, 1, 1, 3200.00, 10.00, 2, 1.00, 1, 'Supplier utama Indomie', '2026-04-17 06:10:11', '2026-04-17 06:10:11'),
(2, 1, 4, 0, 2, 3150.00, 20.00, 3, 1.00, 1, 'Supplier alternatif', '2026-04-17 06:10:11', '2026-04-17 06:10:11'),
(3, 2, 2, 1, 1, 8500.00, 12.00, 3, 1.00, 1, 'Supplier utama sabun', '2026-04-17 06:10:11', '2026-04-17 06:10:11'),
(4, 3, 3, 1, 1, 5000.00, 24.00, 2, 1.00, 1, 'Supplier utama tisu', '2026-04-17 06:10:11', '2026-04-17 06:10:11');

-- --------------------------------------------------------

--
-- Table structure for table `purchase_orders`
--

CREATE TABLE `purchase_orders` (
  `id` bigint(20) NOT NULL,
  `po_number` varchar(50) NOT NULL,
  `stock_check_session_id` bigint(20) DEFAULT NULL,
  `store_id` int(11) NOT NULL,
  `supplier_id` int(11) NOT NULL,
  `po_date` date NOT NULL,
  `status` enum('draft','submitted','sent','received','closed','cancelled') NOT NULL DEFAULT 'draft',
  `created_by` int(11) NOT NULL,
  `notes` text DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `updated_at` timestamp NULL DEFAULT current_timestamp() ON UPDATE current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

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
-- Table structure for table `stock_check_sessions`
--

CREATE TABLE `stock_check_sessions` (
  `id` bigint(20) NOT NULL,
  `session_number` varchar(50) NOT NULL,
  `session_date` date NOT NULL,
  `store_id` int(11) NOT NULL,
  `supplier_id` int(11) NOT NULL,
  `initiation_type` enum('scheduled','checker_initiative') NOT NULL DEFAULT 'scheduled',
  `status` enum('draft','in_progress','submitted','reviewed','closed','cancelled') NOT NULL DEFAULT 'draft',
  `created_by` int(11) NOT NULL,
  `notes` text DEFAULT NULL,
  `submitted_at` datetime DEFAULT NULL,
  `reviewed_at` datetime DEFAULT NULL,
  `closed_at` datetime DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `updated_at` timestamp NULL DEFAULT current_timestamp() ON UPDATE current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `stock_check_sessions`
--

INSERT INTO `stock_check_sessions` (`id`, `session_number`, `session_date`, `store_id`, `supplier_id`, `initiation_type`, `status`, `created_by`, `notes`, `submitted_at`, `reviewed_at`, `closed_at`, `created_at`, `updated_at`) VALUES
(1, 'SCS-MK1-20260418-001', '2026-04-18', 1, 1, 'scheduled', 'in_progress', 1, 'Stock opname supplier ABC Food untuk MK1', NULL, NULL, NULL, '2026-04-18 04:52:25', '2026-04-18 04:52:25');

-- --------------------------------------------------------

--
-- Table structure for table `stock_check_session_items`
--

CREATE TABLE `stock_check_session_items` (
  `id` bigint(20) NOT NULL,
  `stock_check_session_id` bigint(20) NOT NULL,
  `product_id` int(11) NOT NULL,
  `system_qty_store` decimal(18,2) NOT NULL DEFAULT 0.00,
  `system_qty_warehouse` decimal(18,2) NOT NULL DEFAULT 0.00,
  `qty_store` decimal(18,2) NOT NULL DEFAULT 0.00,
  `qty_warehouse` decimal(18,2) NOT NULL DEFAULT 0.00,
  `total_qty` decimal(18,2) NOT NULL DEFAULT 0.00,
  `suggest_buy_qty` decimal(18,2) NOT NULL DEFAULT 0.00,
  `approved_buy_qty` decimal(18,2) DEFAULT NULL,
  `suggested_supplier_id` int(11) DEFAULT NULL,
  `approved_supplier_id` int(11) DEFAULT NULL,
  `condition_status` enum('good','empty_rack','damaged','missing','overstock','other') NOT NULL DEFAULT 'good',
  `status` enum('draft','submitted','reviewed','approved','rejected','po_created') NOT NULL DEFAULT 'draft',
  `checker_notes` varchar(255) DEFAULT NULL,
  `buyer_notes` varchar(255) DEFAULT NULL,
  `reviewed_by` int(11) DEFAULT NULL,
  `reviewed_at` datetime DEFAULT NULL,
  `created_by` int(11) DEFAULT NULL,
  `updated_by` int(11) DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `updated_at` timestamp NULL DEFAULT current_timestamp() ON UPDATE current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `stock_check_session_items`
--

INSERT INTO `stock_check_session_items` (`id`, `stock_check_session_id`, `product_id`, `system_qty_store`, `system_qty_warehouse`, `qty_store`, `qty_warehouse`, `total_qty`, `suggest_buy_qty`, `approved_buy_qty`, `suggested_supplier_id`, `approved_supplier_id`, `condition_status`, `status`, `checker_notes`, `buyer_notes`, `reviewed_by`, `reviewed_at`, `created_by`, `updated_by`, `created_at`, `updated_at`) VALUES
(7, 1, 1, 15.00, 50.00, 12.00, 40.00, 52.00, 20.00, NULL, 1, NULL, 'empty_rack', 'submitted', 'Rak toko mulai kosong', NULL, NULL, NULL, 1, 1, '2026-04-18 05:16:07', '2026-04-18 05:16:07'),
(8, 1, 2, 8.00, 20.00, 7.00, 18.00, 25.00, 15.00, NULL, 1, NULL, 'good', 'submitted', 'Stok menipis', NULL, NULL, NULL, 1, 1, '2026-04-18 05:16:07', '2026-04-18 05:16:07'),
(9, 1, 3, 5.00, 12.00, 4.00, 10.00, 14.00, 25.00, NULL, 1, NULL, 'good', 'submitted', 'Perlu restock', NULL, NULL, NULL, 1, 1, '2026-04-18 05:16:07', '2026-04-18 05:16:07');

-- --------------------------------------------------------

--
-- Table structure for table `stock_check_session_item_histories`
--

CREATE TABLE `stock_check_session_item_histories` (
  `id` bigint(20) NOT NULL,
  `stock_check_session_item_id` bigint(20) NOT NULL,
  `product_id` int(11) NOT NULL,
  `field_name` varchar(100) NOT NULL,
  `old_value` text DEFAULT NULL,
  `new_value` text DEFAULT NULL,
  `change_reason` varchar(255) DEFAULT NULL,
  `notes` text DEFAULT NULL,
  `changed_by` int(11) DEFAULT NULL,
  `changed_at` datetime NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

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
-- Table structure for table `suppliers`
--

CREATE TABLE `suppliers` (
  `id` int(11) NOT NULL,
  `supplier_group_id` int(11) DEFAULT NULL,
  `supplier_code` varchar(50) NOT NULL,
  `supplier_name` varchar(150) NOT NULL,
  `supplier_type` varchar(100) DEFAULT NULL,
  `address` text DEFAULT NULL,
  `phone` varchar(50) DEFAULT NULL,
  `email` varchar(150) DEFAULT NULL,
  `pic_name` varchar(150) DEFAULT NULL,
  `payment_term_days` int(11) DEFAULT 0,
  `is_active` tinyint(1) NOT NULL DEFAULT 1,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `updated_at` timestamp NULL DEFAULT current_timestamp() ON UPDATE current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `suppliers`
--

INSERT INTO `suppliers` (`id`, `supplier_group_id`, `supplier_code`, `supplier_name`, `supplier_type`, `address`, `phone`, `email`, `pic_name`, `payment_term_days`, `is_active`, `created_at`, `updated_at`) VALUES
(1, 1, 'ABC-FOOD', 'ABC Food', 'Food', 'Yogyakarta', '0811111111', NULL, 'Budi', 14, 1, '2026-04-17 04:23:38', '2026-04-17 04:23:38'),
(2, 1, 'ABC-TOY', 'ABC Toiletries', 'Toiletries', 'Yogyakarta', '0822222222', NULL, 'Sari', 14, 1, '2026-04-17 04:23:38', '2026-04-17 04:23:38'),
(3, 1, 'ABC-DEPT', 'ABC Dept Store', 'Dept Store', 'Yogyakarta', '0833333333', NULL, 'Andi', 14, 1, '2026-04-17 04:23:38', '2026-04-17 04:23:38'),
(4, 2, 'XYZ-FOOD', 'XYZ Food', 'Food', 'Sleman', '0844444444', NULL, 'Rina', 7, 1, '2026-04-17 04:23:38', '2026-04-17 04:23:38');

-- --------------------------------------------------------

--
-- Table structure for table `supplier_groups`
--

CREATE TABLE `supplier_groups` (
  `id` int(11) NOT NULL,
  `group_code` varchar(50) NOT NULL,
  `group_name` varchar(150) NOT NULL,
  `description` text DEFAULT NULL,
  `is_active` tinyint(1) NOT NULL DEFAULT 1,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `updated_at` timestamp NULL DEFAULT current_timestamp() ON UPDATE current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `supplier_groups`
--

INSERT INTO `supplier_groups` (`id`, `group_code`, `group_name`, `description`, `is_active`, `created_at`, `updated_at`) VALUES
(1, 'ABC', 'ABC', NULL, 1, '2026-04-17 04:23:38', '2026-04-17 04:23:38'),
(2, 'XYZ', 'XYZ', NULL, 1, '2026-04-17 04:23:38', '2026-04-17 04:23:38');

-- --------------------------------------------------------

--
-- Table structure for table `units`
--

CREATE TABLE `units` (
  `id` int(11) NOT NULL,
  `unit_code` varchar(20) NOT NULL,
  `unit_name` varchar(50) NOT NULL,
  `description` varchar(255) DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `updated_at` timestamp NULL DEFAULT current_timestamp() ON UPDATE current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `units`
--

INSERT INTO `units` (`id`, `unit_code`, `unit_name`, `description`, `created_at`, `updated_at`) VALUES
(1, 'PCS', 'PCS', 'Satuan per pcs', '2026-04-17 04:17:42', '2026-04-17 04:17:42'),
(2, 'BOX', 'BOX', 'Satuan box', '2026-04-17 04:17:42', '2026-04-17 04:17:42'),
(3, 'PACK', 'PACK', 'Satuan pack', '2026-04-17 04:17:42', '2026-04-17 04:17:42'),
(4, 'KRT', 'KARTON', 'Satuan karton', '2026-04-17 04:17:42', '2026-04-17 04:17:42');

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
-- Indexes for table `products`
--
ALTER TABLE `products`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `uk_products_code` (`product_code`),
  ADD UNIQUE KEY `uk_products_barcode` (`barcode`),
  ADD KEY `idx_products_category` (`category_id`),
  ADD KEY `idx_products_unit` (`unit_id`),
  ADD KEY `idx_products_created_by` (`created_by`),
  ADD KEY `idx_products_updated_by` (`updated_by`);

--
-- Indexes for table `product_categories`
--
ALTER TABLE `product_categories`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `uk_product_categories_code` (`category_code`);

--
-- Indexes for table `product_suppliers`
--
ALTER TABLE `product_suppliers`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `uk_product_supplier_unique` (`product_id`,`supplier_id`),
  ADD KEY `idx_product_suppliers_supplier` (`supplier_id`);

--
-- Indexes for table `purchase_orders`
--
ALTER TABLE `purchase_orders`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `uk_purchase_orders_number` (`po_number`);

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
-- Indexes for table `stock_check_sessions`
--
ALTER TABLE `stock_check_sessions`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `uk_stock_check_sessions_number` (`session_number`),
  ADD KEY `idx_stock_check_sessions_date` (`session_date`),
  ADD KEY `idx_stock_check_sessions_store` (`store_id`),
  ADD KEY `idx_stock_check_sessions_supplier` (`supplier_id`),
  ADD KEY `idx_stock_check_sessions_created_by` (`created_by`);

--
-- Indexes for table `stock_check_session_items`
--
ALTER TABLE `stock_check_session_items`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `uk_stock_check_session_item` (`stock_check_session_id`,`product_id`),
  ADD KEY `idx_scsi_product` (`product_id`),
  ADD KEY `idx_scsi_created_by` (`created_by`),
  ADD KEY `idx_scsi_updated_by` (`updated_by`),
  ADD KEY `idx_scsi_reviewed_by` (`reviewed_by`),
  ADD KEY `idx_scsi_suggested_supplier` (`suggested_supplier_id`),
  ADD KEY `idx_scsi_approved_supplier` (`approved_supplier_id`);

--
-- Indexes for table `stock_check_session_item_histories`
--
ALTER TABLE `stock_check_session_item_histories`
  ADD PRIMARY KEY (`id`),
  ADD KEY `idx_scsih_item` (`stock_check_session_item_id`),
  ADD KEY `idx_scsih_product` (`product_id`),
  ADD KEY `idx_scsih_changed_by` (`changed_by`);

--
-- Indexes for table `stores`
--
ALTER TABLE `stores`
  ADD PRIMARY KEY (`store_id`),
  ADD KEY `store_id` (`store_id`);

--
-- Indexes for table `suppliers`
--
ALTER TABLE `suppliers`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `uk_suppliers_code` (`supplier_code`),
  ADD KEY `idx_suppliers_group` (`supplier_group_id`);

--
-- Indexes for table `supplier_groups`
--
ALTER TABLE `supplier_groups`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `uk_supplier_groups_code` (`group_code`);

--
-- Indexes for table `units`
--
ALTER TABLE `units`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `uk_units_code` (`unit_code`);

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
-- AUTO_INCREMENT for table `products`
--
ALTER TABLE `products`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=4;

--
-- AUTO_INCREMENT for table `product_categories`
--
ALTER TABLE `product_categories`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=4;

--
-- AUTO_INCREMENT for table `product_suppliers`
--
ALTER TABLE `product_suppliers`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=5;

--
-- AUTO_INCREMENT for table `purchase_orders`
--
ALTER TABLE `purchase_orders`
  MODIFY `id` bigint(20) NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `roles`
--
ALTER TABLE `roles`
  MODIFY `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=9;

--
-- AUTO_INCREMENT for table `stock_check_sessions`
--
ALTER TABLE `stock_check_sessions`
  MODIFY `id` bigint(20) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=2;

--
-- AUTO_INCREMENT for table `stock_check_session_items`
--
ALTER TABLE `stock_check_session_items`
  MODIFY `id` bigint(20) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=10;

--
-- AUTO_INCREMENT for table `stock_check_session_item_histories`
--
ALTER TABLE `stock_check_session_item_histories`
  MODIFY `id` bigint(20) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=8;

--
-- AUTO_INCREMENT for table `suppliers`
--
ALTER TABLE `suppliers`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=5;

--
-- AUTO_INCREMENT for table `supplier_groups`
--
ALTER TABLE `supplier_groups`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=3;

--
-- AUTO_INCREMENT for table `units`
--
ALTER TABLE `units`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=5;

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
-- Constraints for table `products`
--
ALTER TABLE `products`
  ADD CONSTRAINT `fk_products_category` FOREIGN KEY (`category_id`) REFERENCES `product_categories` (`id`) ON DELETE SET NULL ON UPDATE CASCADE,
  ADD CONSTRAINT `fk_products_created_by` FOREIGN KEY (`created_by`) REFERENCES `users` (`id`) ON DELETE SET NULL ON UPDATE CASCADE,
  ADD CONSTRAINT `fk_products_unit` FOREIGN KEY (`unit_id`) REFERENCES `units` (`id`) ON DELETE SET NULL ON UPDATE CASCADE,
  ADD CONSTRAINT `fk_products_updated_by` FOREIGN KEY (`updated_by`) REFERENCES `users` (`id`) ON DELETE SET NULL ON UPDATE CASCADE;

--
-- Constraints for table `product_suppliers`
--
ALTER TABLE `product_suppliers`
  ADD CONSTRAINT `fk_product_suppliers_product` FOREIGN KEY (`product_id`) REFERENCES `products` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  ADD CONSTRAINT `fk_product_suppliers_supplier` FOREIGN KEY (`supplier_id`) REFERENCES `suppliers` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

--
-- Constraints for table `role_has_permissions`
--
ALTER TABLE `role_has_permissions`
  ADD CONSTRAINT `role_has_permissions_ibfk_1` FOREIGN KEY (`permission_id`) REFERENCES `permissions` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `role_has_permissions_ibfk_2` FOREIGN KEY (`role_id`) REFERENCES `roles` (`id`) ON DELETE CASCADE;

--
-- Constraints for table `stock_check_sessions`
--
ALTER TABLE `stock_check_sessions`
  ADD CONSTRAINT `fk_stock_check_sessions_created_by` FOREIGN KEY (`created_by`) REFERENCES `users` (`id`) ON UPDATE CASCADE,
  ADD CONSTRAINT `fk_stock_check_sessions_store` FOREIGN KEY (`store_id`) REFERENCES `stores` (`store_id`) ON UPDATE CASCADE,
  ADD CONSTRAINT `fk_stock_check_sessions_supplier` FOREIGN KEY (`supplier_id`) REFERENCES `suppliers` (`id`) ON UPDATE CASCADE;

--
-- Constraints for table `stock_check_session_items`
--
ALTER TABLE `stock_check_session_items`
  ADD CONSTRAINT `fk_scsi_approved_supplier` FOREIGN KEY (`approved_supplier_id`) REFERENCES `suppliers` (`id`) ON DELETE SET NULL ON UPDATE CASCADE,
  ADD CONSTRAINT `fk_scsi_created_by` FOREIGN KEY (`created_by`) REFERENCES `users` (`id`) ON DELETE SET NULL ON UPDATE CASCADE,
  ADD CONSTRAINT `fk_scsi_product` FOREIGN KEY (`product_id`) REFERENCES `products` (`id`) ON UPDATE CASCADE,
  ADD CONSTRAINT `fk_scsi_reviewed_by` FOREIGN KEY (`reviewed_by`) REFERENCES `users` (`id`) ON DELETE SET NULL ON UPDATE CASCADE,
  ADD CONSTRAINT `fk_scsi_session` FOREIGN KEY (`stock_check_session_id`) REFERENCES `stock_check_sessions` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  ADD CONSTRAINT `fk_scsi_suggested_supplier` FOREIGN KEY (`suggested_supplier_id`) REFERENCES `suppliers` (`id`) ON DELETE SET NULL ON UPDATE CASCADE,
  ADD CONSTRAINT `fk_scsi_updated_by` FOREIGN KEY (`updated_by`) REFERENCES `users` (`id`) ON DELETE SET NULL ON UPDATE CASCADE;

--
-- Constraints for table `stock_check_session_item_histories`
--
ALTER TABLE `stock_check_session_item_histories`
  ADD CONSTRAINT `fk_scsih_changed_by` FOREIGN KEY (`changed_by`) REFERENCES `users` (`id`) ON DELETE SET NULL ON UPDATE CASCADE,
  ADD CONSTRAINT `fk_scsih_item` FOREIGN KEY (`stock_check_session_item_id`) REFERENCES `stock_check_session_items` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  ADD CONSTRAINT `fk_scsih_product` FOREIGN KEY (`product_id`) REFERENCES `products` (`id`) ON UPDATE CASCADE;

--
-- Constraints for table `suppliers`
--
ALTER TABLE `suppliers`
  ADD CONSTRAINT `fk_suppliers_group` FOREIGN KEY (`supplier_group_id`) REFERENCES `supplier_groups` (`id`) ON DELETE SET NULL ON UPDATE CASCADE;

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

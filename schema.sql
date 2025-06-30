-- MySQL dump 10.13  Distrib 9.3.0, for Linux (x86_64)
--
-- Host: localhost    Database: test
-- ------------------------------------------------------
-- Server version	9.3.0

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `chunk_metadata`
--

DROP TABLE IF EXISTS `chunk_metadata`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `chunk_metadata` (
  `file_chunk_id` int NOT NULL AUTO_INCREMENT,
  `file_chunk_name` varchar(255) DEFAULT NULL,
  `iteration` int DEFAULT NULL,
  `file_name` varchar(30) DEFAULT NULL,
  PRIMARY KEY (`file_chunk_id`),
  UNIQUE KEY `file_chunk_name` (`file_chunk_name`),
  KEY `fk_file_name` (`file_name`),
  CONSTRAINT `fk_file_name` FOREIGN KEY (`file_name`) REFERENCES `fileinfo` (`filename`)
) ENGINE=InnoDB AUTO_INCREMENT=401 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `file_metadata`
--

DROP TABLE IF EXISTS `file_metadata`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `file_metadata` (
  `id` int NOT NULL AUTO_INCREMENT,
  `file_name` varchar(255) DEFAULT NULL,
  `file_chunk_name` varchar(255) DEFAULT NULL,
  `k0_name` varchar(255) DEFAULT NULL,
  `k1_name` varchar(255) DEFAULT NULL,
  `k2_name` varchar(255) DEFAULT NULL,
  `r1_name` varchar(255) DEFAULT NULL,
  `r2_name` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `k0_name` (`k0_name`),
  UNIQUE KEY `k1_name` (`k1_name`),
  UNIQUE KEY `k2_name` (`k2_name`),
  UNIQUE KEY `r1_name` (`r1_name`),
  UNIQUE KEY `r2_name` (`r2_name`),
  KEY `file_name` (`file_name`),
  KEY `file_chunk_name` (`file_chunk_name`),
  CONSTRAINT `file_metadata_ibfk_1` FOREIGN KEY (`file_name`) REFERENCES `fileinfo` (`filename`) ON DELETE CASCADE,
  CONSTRAINT `file_metadata_ibfk_2` FOREIGN KEY (`file_chunk_name`) REFERENCES `chunk_metadata` (`file_chunk_name`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fileinfo`
--

DROP TABLE IF EXISTS `fileinfo`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `fileinfo` (
  `id` int NOT NULL AUTO_INCREMENT,
  `filename` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `filename` (`filename`)
) ENGINE=InnoDB AUTO_INCREMENT=19 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2025-06-30  0:15:13

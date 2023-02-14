CREATE PROCEDURE drop_venus_database()
BEGIN
    DECLARE done INT DEFAULT FALSE;
    DECLARE dbname VARCHAR(255);
    DECLARE cur CURSOR FOR SELECT schema_name
                           FROM information_schema.schemata
                           WHERE schema_name LIKE 'venus%'
                           ORDER BY schema_name;
DECLARE CONTINUE HANDLER FOR NOT FOUND SET done = TRUE;

OPEN cur;

read_loop: LOOP
        FETCH cur INTO dbname;

        IF done THEN
          LEAVE read_loop;
END IF;

        SET @query = CONCAT('DROP DATABASE `',dbname, '`');
PREPARE stmt FROM @query;
EXECUTE stmt;
END LOOP;
END;

CALL drop_venus_database();

DROP PROCEDURE drop_venus_database


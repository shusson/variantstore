CREATE DATABASE variants;
USE variants;

CREATE TABLE mgrb (
		VARIANT TEXT,
		CHROMOSOME VARCHAR(2),
		START INT,
		REF TEXT,
		ALT TEXT,
		RSID TEXT,
		AC INT,
		AF DECIMAL(6,5),
    nHomRef INT,
    nHet INT,
    nHomVar INT,
		TYPE VARCHAR(32),
		CATO DECIMAL(6,5),
		eigen DECIMAL(6,5),
		sift TEXT,
    HGVSc TEXT,
    HGVSp TEXT,
		polyPhen TEXT,
		tgpAF DECIMAL(6,5),
		hrcAF DECIMAL(6,5),
		gnomadAF DECIMAL(6,5),
		feature TEXT,
		consequences TEXT,
		gene TEXT,
		clinvar TEXT,
		INDEX name (chromosome, start)
	);

LOAD DATA INFILE '/data/sample.tsv' INTO TABLE mgrb
				  FIELDS TERMINATED BY '\t'
				  LINES TERMINATED BY '\n'
				  IGNORE 1 LINES;
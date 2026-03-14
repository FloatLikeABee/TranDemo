WITH EthnicCodes AS (
    SELECT
        oec.stud_id,
        ec.dbid,
        STRING_AGG(CAST(ec.Code AS VARCHAR(MAX)), ', ') AS EthnicCode
    FROM
        dbo.Ethnic_Codes ec
    INNER JOIN
        dbo.OwnedEthnicCodes oec ON ec.EthnicCodeID = oec.EthnicCodeID AND ec.DBID = oec.DBID
    GROUP BY
        ec.dbid, oec.stud_id
),
DisabilityCodes AS (
    SELECT
        od.stud_id,
        od.dbid,
        STRING_AGG(CAST(dc.Code AS VARCHAR(MAX)), ', ') AS DisabilityCodes
    FROM
        dbo.OwnedDisability od
    INNER JOIN
        dbo.disability_codes dc ON od.DisCodeID = dc.DisCodeID AND od.DBID = dc.DBID
    GROUP BY
        od.dbid, od.stud_id
)

SELECT
    Student.Stud_ID,
    Student.Last_Name,
    Student.First_Name,
    Student.Dob,
    DATEDIFF(YEAR, Student.Dob, GETDATE()) AS Age,
    Gender.Code AS Gender,
    Student.Email,
    Student.Geo_City,
    Student.Geo_Street,
    Grade.Code AS Grade,
    district.district AS District,
    school.Name AS SchoolName,
    Student.Disabled,
    Student.Transported,
    EthnicCodes.EthnicCode,
    DisabilityCodes.DisabilityCodes
FROM
    dbo.Student
LEFT JOIN
    dbo.grade ON grade.ID = student.grade
LEFT JOIN
    dbo.Gender ON Gender.ID = student.Gender
LEFT JOIN
    dbo.district ON STUDENT.DistrictID = district.DistrictID AND district.DBID = Student.DBID
LEFT JOIN
    dbo.school ON STUDENT.school = school.SchoolCode AND school.DBID = Student.DBID
LEFT JOIN
    EthnicCodes ON EthnicCodes.DBID = Student.DBID AND EthnicCodes.Stud_ID = Student.Stud_ID
LEFT JOIN
    DisabilityCodes ON DisabilityCodes.DBID = Student.DBID AND DisabilityCodes.Stud_ID = Student.Stud_ID;

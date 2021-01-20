# PathwayScore
     Software to calculate Rank based pathway score

Software requires expression matrix as csv file, and pathway file as txt file

##### formate of expn data in csv file
sample	| A4GALT |	AAAS |	AACS
------- | ------ | ----- | ------ 
Sample_01 |	3.124949 |	3.26064 | 2.369848
Sample_02 |	2.999860 |	3.11805	| 1.579720
Sample_03 |	2.318580 |	3.56156	| 2.875882

by default software use max 2 cores but it can be reset with --nCPU parameter (eg __--nCPU 4__)

* ## Usage
    * ##### build first and run the programme or just run the programme
    * #### build first
        * $ go build parRankScore.go
    * #### run 
        * $ ./parRankScore.go --filename "Project_data_expn_matrix.csv" --Pathway "wiki" --nCPU 6
    * ##### build and run at once
        * $ go run ./parRankScore.go --filename "Project_data_expn_matrix.csv" --Pathway "wiki" --nCPU 6




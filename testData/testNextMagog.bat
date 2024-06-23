:: Each opening as black and white against each player. After playing the opening with every player go to next opening. 
:: Set number of rounds to number of openings!

IF [%1] == [] ( 
	ECHO Tournament name missing
	goto :eof
)

cutechess-cli -tournament gauntlet -openings file=basicOpenings.pgn policy=round -debug ^
-resign movecount=5 score=1400 twosided=true ^
-maxmoves 150 ^
-concurrency 2 ^
-pgnout %1"games.pgn" -event %1 ^
-engine conf=magog.Challenger ^
-engine conf=magog.Defender ^
-engine conf=DoctorB ^
-engine conf=EnkoChess_290818 ^
-each tc=120+1 -games 2 -rounds 7 -repeat > %1".debug"
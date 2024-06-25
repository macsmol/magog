# Magog Chess Engine

UCI chess engine written from scratch in Golang. Started as a competitor to my previous engine done in Java - Machess.
But generally it's just an excuse to learn some Go. 

Currently, it plays with modest strength and I expect it to be around 1350-1450 ELO on CCRL blitz list.

Feel free to try it out! Please report any obvious bugs or crashes you may find. :)

## Features

### Search
* Alpha-beta search with iterative deepening
* Quiescence search
* Move ordering
  * PV-move
  * Captures/promotions according to material difference
  * killer moves

### Board representation
* 0x88 board
* piece lists (3 per player): king, pawns, other pieces

### Evaluation
* Lazy evaluation between:
  * Tapered variant of Simplified Evaluation Function
  * Tapered variant of Simplified Evaluation Function + mobility

### Non-Uci commands
* `perft <depth>` - count number of moves possible from current position
* `tperft <depth>` - same as perft but at `<depth>` count only captures and promotions. Useful for testing movegen in quiescence search.
* `tostr` - print board representation of current position

## Compilation
To build *.exe file run this in repository root: 
>go build .

To run perft test suite run:
>go test .\engine -v

## Credits
This engine would not have been possible if it weren't for many people who shared their knowledge:
* [Chess Programming Wiki](https://www.chessprogramming.org/)
* [Articles by Bruce Moreland](https://web.archive.org/web/20070811182741/http://www.seanet.com/~brucemo/topics/topics.htm)
* [Bluefever Software series on YT](https://www.youtube.com/watch?v=bGAfaepBco4&list=PLZ1QII7yudbc-Ky058TEaOstZHVbT-2hg)
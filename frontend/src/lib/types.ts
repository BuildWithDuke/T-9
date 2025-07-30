export enum Player {
  Empty = 0,
  X = 1,
  O = 2
}

export type SmallBoard = Player[];
export type BigBoard = SmallBoard[];
export type BigBoardWins = Player[];

export interface GameState {
  bigBoard: BigBoard;
  bigBoardWins: BigBoardWins;
  activeBoard: number; // -1 means can play anywhere
  currentPlayer: Player;
  gameWon: Player;
  gameOver: boolean;
}

export interface Move {
  bigBoardIndex: number;
  smallBoardIndex: number;
  player: Player;
}

export interface GameResponse {
  gameId: string;
  game: GameState;
}
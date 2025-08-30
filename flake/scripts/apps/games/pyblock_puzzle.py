#!/usr/bin/env python3

import random
import time
import os

# A Tetris-like game written in Python.
# This version includes a full-featured mode using the 'curses' library
# and a fallback text-based mode for maximum compatibility.

# Attempt to import curses and set a flag.
try:
    import curses
    CURSES_AVAILABLE = True
except ImportError:
    CURSES_AVAILABLE = False

class TetrisGame:
    """A class to encapsulate the Tetris game logic and state."""

    def __init__(self):
        # --- Game Configuration ---
        self.W, self.H = 10, 20
        self.PIECE_CHAR = "■"
        self.GAME_LOOP_SLEEP = 0.01 # Small sleep to prevent 100% CPU usage

        # Tetromino definitions (4x4 blocks)
        self.TETROMINOS = [
            [[0, 1, 0, 0], [0, 1, 0, 0], [0, 1, 0, 0], [0, 1, 0, 0]],  # I
            [[0, 1, 0], [1, 1, 1], [0, 0, 0]],                      # T
            [[1, 1], [1, 1]],                                      # O
            [[0, 1, 1], [1, 1, 0], [0, 0, 0]],                      # Z
            [[1, 1, 0], [0, 1, 1], [0, 0, 0]],                      # S
            [[1, 0, 0], [1, 1, 1], [0, 0, 0]],                      # J
            [[0, 0, 1], [1, 1, 1], [0, 0, 0]],                      # L
        ]
        self.COLORS = [1, 2, 3, 4, 5, 6, 7]
        self.TEXT_COLORS = {
            1: '\033[96m', 2: '\033[95m', 3: '\033[93m', 4: '\033[91m',
            5: '\033[92m', 6: '\033[94m', 7: '\033[37m', 'border': '\033[90m',
            'reset': '\033[0m'
        }

        # --- Game State ---
        self.board = [[0 for _ in range(self.W)] for _ in range(self.H)]
        self.score = 0
        self.level = 1
        self.lines_cleared = 0
        self.game_over = False
        self.current_piece = None
        self.next_piece = None

    # --- Core Game Logic Methods ---
    def new_piece(self):
        shape = random.choice(self.TETROMINOS)
        color = random.choice(self.COLORS)
        return {'shape': shape, 'color': color, 'x': self.W // 2 - len(shape[0]) // 2, 'y': 0}

    def rotate_piece(self, piece):
        shape = piece['shape']
        rotated = [list(row) for row in zip(*shape[::-1])]
        return {**piece, 'shape': rotated}

    def check_collision(self, piece, offset_x=0, offset_y=0):
        for y, row in enumerate(piece['shape']):
            for x, cell in enumerate(row):
                if cell:
                    board_x = piece['x'] + x + offset_x
                    board_y = piece['y'] + y + offset_y
                    if not (0 <= board_x < self.W and 0 <= board_y < self.H and self.board[board_y][board_x] == 0):
                        return True
        return False

    def lock_piece(self):
        for y, row in enumerate(self.current_piece['shape']):
            for x, cell in enumerate(row):
                if cell:
                    self.board[self.current_piece['y'] + y][self.current_piece['x'] + x] = self.current_piece['color']
        
        lines_to_clear = [i for i, row in enumerate(self.board) if all(row)]
        if lines_to_clear:
            for i in lines_to_clear:
                del self.board[i]
                self.board.insert(0, [0 for _ in range(self.W)])
            
            cleared_count = len(lines_to_clear)
            self.lines_cleared += cleared_count
            self.score += [0, 40, 100, 300, 1200][cleared_count] * self.level
            self.level = 1 + self.lines_cleared // 10

    # --- Curses Mode (Full Experience) ---
    def run_curses(self, stdscr):
        curses.curs_set(0)
        curses.start_color()
        curses.init_pair(1, curses.COLOR_CYAN, curses.COLOR_BLACK)
        curses.init_pair(2, curses.COLOR_MAGENTA, curses.COLOR_BLACK)
        curses.init_pair(3, curses.COLOR_YELLOW, curses.COLOR_BLACK)
        curses.init_pair(4, curses.COLOR_RED, curses.COLOR_BLACK)
        curses.init_pair(5, curses.COLOR_GREEN, curses.COLOR_BLACK)
        curses.init_pair(6, curses.COLOR_BLUE, curses.COLOR_BLACK)
        curses.init_pair(7, curses.COLOR_WHITE, curses.COLOR_BLACK)
        stdscr.nodelay(True)

        self.current_piece = self.new_piece()
        self.next_piece = self.new_piece()
        last_fall_time = time.time()

        while not self.game_over:
            key = stdscr.getch()
            if key != -1:
                if key == ord('q'): break
                elif key == ord('a'):
                    if not self.check_collision(self.current_piece, offset_x=-1): self.current_piece['x'] -= 1
                elif key == ord('d'):
                    if not self.check_collision(self.current_piece, offset_x=1): self.current_piece['x'] += 1
                elif key == ord('s'):
                    if not self.check_collision(self.current_piece, offset_y=1):
                        self.current_piece['y'] += 1
                        last_fall_time = time.time()
                elif key == ord('w'):
                    rotated = self.rotate_piece(self.current_piece)
                    if not self.check_collision(rotated): self.current_piece = rotated

            fall_speed = max(0.1, 0.5 - (self.level - 1) * 0.05)
            if time.time() - last_fall_time > fall_speed:
                if not self.check_collision(self.current_piece, offset_y=1):
                    self.current_piece['y'] += 1
                else:
                    self.lock_piece()
                    self.current_piece = self.next_piece
                    self.next_piece = self.new_piece()
                    if self.check_collision(self.current_piece): self.game_over = True
                last_fall_time = time.time()

            stdscr.clear()
            # --- DRAWING ---
            stdscr.addstr(0, 0, f"Score: {self.score}  Level: {self.level}  Lines: {self.lines_cleared}")
            
            for y, row in enumerate(self.board):
                for x, cell in enumerate(row):
                    if cell: stdscr.addstr(y + 2, x * 2 + 2, self.PIECE_CHAR * 2, curses.color_pair(cell))
            for y, row in enumerate(self.current_piece['shape']):
                for x, cell in enumerate(row):
                    if cell: stdscr.addstr(self.current_piece['y'] + y + 2, (self.current_piece['x'] + x) * 2 + 2, self.PIECE_CHAR * 2, curses.color_pair(self.current_piece['color']))
            
            for y in range(self.H + 2): 
                stdscr.addstr(y + 1, 0, "■")
                stdscr.addstr(y + 1, self.W * 2 + 2, "■")
            for x in range(self.W * 2 + 3): 
                stdscr.addstr(1, x, "■")
                stdscr.addstr(self.H + 2, x, "■")

            stdscr.addstr(2, self.W * 2 + 6, "Next:")
            for y, row in enumerate(self.next_piece['shape']):
                for x, cell in enumerate(row):
                    if cell: stdscr.addstr(y + 4, (x + self.W) * 2 + 8, self.PIECE_CHAR * 2, curses.color_pair(self.next_piece['color']))

            stdscr.refresh()
            time.sleep(self.GAME_LOOP_SLEEP)
        
        stdscr.nodelay(False)
        stdscr.clear()
        stdscr.addstr(self.H // 2, self.W - 4, "GAME OVER")
        stdscr.addstr(self.H // 2 + 2, self.W - 10, "Press any key to exit")
        stdscr.getch()
    
    # --- Text Mode (Fallback) ---
    def run_text(self):
        self.current_piece = self.new_piece()
        self.next_piece = self.new_piece()

        while not self.game_over:
            os.system('clear' if os.name == 'posix' else 'cls')
            
            temp_board = [row[:] for row in self.board]
            for y, row in enumerate(self.current_piece['shape']):
                for x, cell in enumerate(row):
                    if cell and 0 <= self.current_piece['y'] + y < self.H and 0 <= self.current_piece['x'] + x < self.W:
                        temp_board[self.current_piece['y'] + y][self.current_piece['x'] + x] = self.current_piece['color']
            
            print(f"Score: {self.score}  Level: {self.level}  Lines: {self.lines_cleared}")
            print(self.TEXT_COLORS['border'] + "■" * (self.W * 2 + 2) + self.TEXT_COLORS['reset'])
            for y, row in enumerate(temp_board):
                line = self.TEXT_COLORS['border'] + "■ " + self.TEXT_COLORS['reset']
                for cell in row:
                    if cell:
                        line += f"{self.TEXT_COLORS[cell]}{self.PIECE_CHAR} {self.TEXT_COLORS['reset']}"
                    else:
                        line += "  "
                line += self.TEXT_COLORS['border'] + "■" + self.TEXT_COLORS['reset']
                print(line)
            print(self.TEXT_COLORS['border'] + "■" * (self.W * 2 + 2) + self.TEXT_COLORS['reset'])

            key = input("Move (a,d,w,s) or q to quit: ")
            if key == 'q': break
            elif key == 'a':
                if not self.check_collision(self.current_piece, offset_x=-1): self.current_piece['x'] -= 1
            elif key == 'd':
                if not self.check_collision(self.current_piece, offset_x=1): self.current_piece['x'] += 1
            elif key == 'w':
                rotated = self.rotate_piece(self.current_piece)
                if not self.check_collision(rotated): self.current_piece = rotated
            
            if not self.check_collision(self.current_piece, offset_y=1):
                self.current_piece['y'] += 1
            else:
                self.lock_piece()
                self.current_piece = self.next_piece
                self.next_piece = self.new_piece()
                if self.check_collision(self.current_piece): self.game_over = True
        
        print("\nGAME OVER")
        print(f"Final Score: {self.score}, Level: {self.level}")

    # --- Main Execution ---
    def run(self):
        if CURSES_AVAILABLE:
            try:
                curses.wrapper(self.run_curses)
            except curses.error as e:
                os.system('clear' if os.name == 'posix' else 'cls')
                print("="*50)
                print(" A graphical (curses) error occurred.")
                print(f" Error: {e}")
                print(" This usually means your terminal is not fully compatible.")
                print("="*50)
                print("\nFalling back to simple text mode...")
                time.sleep(4)
                self.run_text()
        else:
            print("Python 'curses' library not found.")
            print("Falling back to simple text mode.")
            time.sleep(3)
            self.run_text()

if __name__ == "__main__":
    game = TetrisGame()
    try:
        game.run()
    finally:
        print("\nGame has ended.")
        input("Press Enter to exit.")

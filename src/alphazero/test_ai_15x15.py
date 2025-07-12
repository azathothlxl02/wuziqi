
import numpy as np
import pickle
import sys
import os

sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..'))

from game import Game, Board
from mcts_alphaZero import MCTSPlayer
from policy_value_net_numpy import PolicyValueNetNumpy

# ---------- 1. 常量 ----------
BOARD_SIZE = 8
N_IN_ROW   = 5
MODEL_FILE = 'best_policy_8_8_5.model2'   # 保证该文件存在

# ---------- 2. 加载模型 ----------
try:
    policy_param = pickle.load(open(MODEL_FILE, 'rb'))
except Exception:
    policy_param = pickle.load(open(MODEL_FILE, 'rb'), encoding='bytes')

net  = PolicyValueNetNumpy(BOARD_SIZE, BOARD_SIZE, policy_param)
ai   = MCTSPlayer(net.policy_value_fn, c_puct=5, n_playout=400)

# ---------- 3. 生成随机棋盘 ----------
def random_board():
    # 随机比例：空 70%，黑 15%，白 15%
    rnd = np.random.choice([0, 1, 2], size=BOARD_SIZE*BOARD_SIZE, p=[0.7, 0.15, 0.15])
    return rnd.tolist()

flat = random_board()

# ---------- 4. 构造 Game ----------
board = Board(width=BOARD_SIZE, height=BOARD_SIZE, n_in_row=N_IN_ROW)
board.init_board(start_player=0)
for idx, val in enumerate(flat):
    if val != 0:
        board.states[idx] = int(val)
game = Game(board)


# 固定让 AI 下黑方（1）
game.current_player = 1

# ---------- 5. 调用 get_action ----------
move_flat = ai.get_action(game.board)
row, col  = divmod(move_flat, BOARD_SIZE)

# ---------- 6. 打印结果 ----------
print("Flat move :", move_flat)
print("Row, Col  :", row, col)

# 简单可视化
board_2d = np.array(flat).reshape(BOARD_SIZE, BOARD_SIZE)
print("\nRandom board (0=empty, 1=black, 2=white):")
print(board_2d)

# 把 AI 落子标出来
board_2d[row, col] = 8   # 用 8 高亮
print("\nAfter AI move (8 = choice):")
print(board_2d)
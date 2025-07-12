#!/usr/bin/env python3
"""
Go → 15×15 五子棋 → PyTorch 模型 → 下一步坐标
输入 JSON: {"board":[0,1,2,...]}  0空 1黑 2白
输出 JSON: {"row":r,"col":c}
"""
import sys
import json
import numpy as np
import torch

sys.path.insert(0, '.')
from alphazero.game import Board
from alphazero.mcts_alphaZero import MCTSPlayer
from alphazero.policy_value_net_pytorch import PolicyValueNet

BOARD_SIZE = 15
N_IN_ROW   = 5
MODEL_FILE = 'alphazero/best_policy_8_8_5.model'

# 1. 加载网络
net = PolicyValueNet(BOARD_SIZE, BOARD_SIZE,
                     model_file=MODEL_FILE,
                     use_gpu=torch.cuda.is_available())
mcts_player = MCTSPlayer(net.policy_value_fn,
                         c_puct=5,
                         n_playout=400)

# 2. 推理
def predict(req):
    flat = req["board"]
    board = Board(width=BOARD_SIZE, height=BOARD_SIZE, n_in_row=N_IN_ROW)
    board.init_board(start_player=0)          # 确保 availables 存在
    for idx, val in enumerate(flat):
        if val != 0:
            board.states[idx] = int(val)
    # 固定让 AI 下黑方（1）
    board.current_player = 1
    move = mcts_player.get_action(board)
    row, col = divmod(move, BOARD_SIZE)
    return {"row": row, "col": col}

if __name__ == "__main__":
    req = json.load(sys.stdin)
    print(json.dumps(predict(req)), flush=True)
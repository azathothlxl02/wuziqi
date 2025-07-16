
import sys
import json
import pickle
import numpy as np

sys.path.insert(0, '.')

from alphazero.game import Board
from alphazero.mcts_alphaZero import MCTSPlayer
from alphazero.policy_value_net_numpy import PolicyValueNetNumpy

BOARD_SIZE = 8
N_IN_ROW   = 5
MODEL_FILE = './src/alphazero/best_policy_8_8_5.model'  

try:
    with open(MODEL_FILE, 'rb') as f:
        params = pickle.load(f)
except:
    with open(MODEL_FILE, 'rb') as f:
        params = pickle.load(f, encoding='bytes')

net  = PolicyValueNetNumpy(BOARD_SIZE, BOARD_SIZE, params)
mcts = MCTSPlayer(net.policy_value_fn, c_puct=5, n_playout=400)

def predict(req):
    flat = req["board"]
    board = Board(width=BOARD_SIZE, height=BOARD_SIZE, n_in_row=N_IN_ROW)
    board.init_board(start_player=0)

    for idx, val in enumerate(flat):
        if val != 0:
            board.states[idx] = int(val)

    board.current_player = 1
    board.availables = [i for i, v in enumerate(flat) if v == 0]
    move = mcts.get_action(board)

    row, col = int(move // BOARD_SIZE), int(move % BOARD_SIZE)
    return {"row": row, "col": col}

if __name__ == "__main__":
    req = json.load(sys.stdin)
    print(json.dumps(predict(req), separators=(',', ':')), flush=True)
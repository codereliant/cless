from flask import Flask
import os
app = Flask(__name__)

@app.route('/')
def hello_world():
    gretting = os.environ.get('GREETING', 'cLess')
    return f'Hello, {gretting}!\n'
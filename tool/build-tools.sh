#!/bin/bash

pyinstaller my_init.py
pyinstaller setuser.py

cp -R dist/setuser/* ../image/tool/py_tools
cp -R dist/my_init/* ../image/tool/py_tools

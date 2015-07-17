#!/bin/bash

rm -rf ../image/tool/py_tools/*

pyinstaller -p /usr/local/lib/python2.7/dist-packages/ my_init.py
pyinstaller -p /usr/local/lib/python2.7/dist-packages/ setuser.py

cp -R dist/setuser/* ../image/tool/py_tools
cp -R dist/my_init/* ../image/tool/py_tools

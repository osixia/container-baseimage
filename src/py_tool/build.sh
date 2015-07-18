#!/bin/bash

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

rm -rf $DIR/../../image/tool/py_tool/*

pyinstaller --distpath=$DIR/dist --workpath=$DIR/build --specpath=$DIR -p /usr/local/lib/python2.7/dist-packages/ $DIR/my_init.py
pyinstaller --distpath=$DIR/dist --workpath=$DIR/build --specpath=$DIR -p /usr/local/lib/python2.7/dist-packages/ $DIR/setuser.py 

cp -R $DIR/dist/setuser/* $DIR/../../image/tool/py_tool
cp -R $DIR/dist/my_init/* $DIR/../../image/tool/py_tool

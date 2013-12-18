#!/bin/bash
echo $1

function a() {
    echo "in func"
    echo $1
}

echo "------------"
a $1
echo "------------"
a

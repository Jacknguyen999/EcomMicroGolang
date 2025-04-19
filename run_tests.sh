#!/bin/bash

# Make the test scripts executable
chmod +x test_recommender.sh
chmod +x test_recommender.py

# Check if Python and required packages are installed
if command -v python3 &>/dev/null; then
    echo "Python 3 is installed"
    
    # Install required packages
    pip install requests colorama
    
    # Run the Python test script
    python3 test_recommender.py
else
    echo "Python 3 is not installed, using Bash script instead"
    
    # Run the Bash test script
    ./test_recommender.sh
fi

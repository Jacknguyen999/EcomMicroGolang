#!/bin/bash

# Replace all occurrences of the old GitHub username with the new one
find . -type f -name "*.go" -exec sed -i 's|github.com/rasadov/EcommerceAPI|github.com/thomas/EcommerceAPI|g' {} \;

echo "Import paths updated successfully!"

#!/bin/bash

# Define the base name for the generated files
base_name="teamproject"

# Loop through numbers 1 to 50
for i in {1..100}
do
 # Generate the YAML file name
 file_name="${base_name}-${i}.yaml"

 # Replace the placeholder in the template file with the current number
 sed "s/{number}/${i}/g" teamproject-template.yaml > "${file_name}"

 # Apply the generated YAML file using kubectl
 # kubectl apply -f "${file_name}"
done


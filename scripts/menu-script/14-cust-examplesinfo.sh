while true; do
    echo
    read -p "-> Add another example file? [y|n] " yn
    case $yn in
        [Yy]* ) 
        ./15-createexfiles.sh
        echo
        break;;
        [Nn]* ) 
        echo
        read -p "File not created. Examples should be added to a file with the path examples/resource_directory/example.tf" ; exit;;
        * ) echo; echo "Please answer yes or no.";;
    esac
done
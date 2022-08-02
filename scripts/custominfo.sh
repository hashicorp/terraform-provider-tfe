read -p "Custom templates are used to add more sections to the basic template."
read -p "The common uses are:

    a) > 1 example usage file
    b) > 1 import statement file
    c) notes or warnings that cannot be included in the schema descriptions (general or attribute-specific)
    d) additional sections or headings"

read -p "If you are unsure, first run through Main Menu > Basic."

while true; do
    echo
    read -p "   -> Generate a custom template file? [y|n] " yn
    case $yn in
        [Yy]* ) 
        # TODO: script to create template file
        echo
        # TODO: add path
        read -p "File created at templates/.... Use markdown to edit."; break;;
        [Nn]* ) 
        echo
        read -p "File not created. Custom templates should be added within the templates directory" ; exit;;
        * ) echo "Please answer yes or no.";;
    esac
done

while true; do
    echo
    read -p "   -> Add additional example files? [y|n] " yn
    case $yn in
        [Yy]* ) 
        # TODO: script to create example file (pass in $filename)



        # TODO: add path
        read -p "File created at templates/..."; break;;
        [Nn]* ) 
        echo
        read -p "File not created. Custom templates should be added within the templates directory" ; exit;;
        * ) echo "Please answer yes or no.";;
    esac
done
echo "Inserting authorized_keys...";
echo "$AUTHORIZED_KEYS" > /mnt/.ssh/authorized_keys;

for name in $SSH_KEY_PAIRS; do
    echo "Inserting key $name...";

    priv_key_env="SSH_KEY_${name}_PRIVATE";
    pub_key_env="SSH_KEY_${name}_PUBLIC";
    priv_key="$(printenv $priv_key_env)";
    pub_key="$(printenv $pub_key_env)";

    echo "$priv_key" > "/mnt/.ssh/$name";
    echo "$pub_key" > "/mnt/.ssh/$name.pub";
done

echo "Insertion done.";

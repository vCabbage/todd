while true
do
sleep 5
todd run test-http -j -y
sleep 5
todd run test-bandwidth -j -y
sleep 5
todd run test-ping -j -y
done


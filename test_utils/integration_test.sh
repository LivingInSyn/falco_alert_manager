python3 gen_events.py 5
EVENTS=$(curl -s "localhost:8081/event?page=0&per=50&includeAcknowledged=false")
EVENT_LEN=$(echo $EVENTS | jq '. | length')

echo $EVENT_LEN
if [[ "$EVENT_LEN" != "5" ]]
then
    echo "WRONG LEN - FAIL"
    exit 1
fi

FIRST_ID=$(echo $EVENTS | jq '.[0].ID')
./ack_event.sh $FIRST_ID

EVENT_LEN_A=$(curl -s "localhost:8081/event?page=0&per=50&includeAcknowledged=false" | jq '. | length')
if [[ $EVENT_LEN_A != "4" ]]
then
    echo "WRONG LEN AFTER ACK - FAIL"
    exit 1
fi

echo "PASS"
exit 0
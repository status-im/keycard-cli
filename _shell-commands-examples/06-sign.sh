keycard-select
keycard-set-secrets 123456 123456789012 KeycardTest
keycard-pair

keycard-open-secure-channel
keycard-verify-pin {{ session_pin }}

keycard-derive-key m/1/2/3
keycard-sign 0000000000000000000000000000000000000000000000000000000000000000

keycard-unpair {{ session_pairing_index }}

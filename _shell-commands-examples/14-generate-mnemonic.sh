keycard-select
keycard-set-secrets 123456 123456789012 KeycardDefaultPairing
keycard-pair

keycard-open-secure-channel

keycard-generate-mnemonic 4
keycard-generate-mnemonic 5
keycard-generate-mnemonic 8

keycard-verify-pin {{ session_pin }}
keycard-unpair {{ session_pairing_index }}

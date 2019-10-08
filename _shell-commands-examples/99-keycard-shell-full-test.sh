# select with raw apdu command
gp-send-apdu 00A4040000

# install
gp-open-secure-channel
gp-delete D2760000850101
gp-delete A00000080400010101
gp-delete A0000008040001
gp-load _assets/keycard_v2.2.1.cap A0000008040001
gp-install-for-install A0000008040001 A000000804000102 D2760000850101 0024d40f12616e64726f69642e636f6d3a706b67696d2e7374617475732e657468657265756d
gp-install-for-install A0000008040001 A000000804000101 A00000080400010101

# init
keycard-select
keycard-init

# pair
keycard-select
keycard-pair
keycard-open-secure-channel

# get status
keycard-get-status

# verify PIN
keycard-verify-pin {{ session_pin }}

# change secrets
keycard-change-pin 888888
keycard-change-puk 111222333444
keycard-change-pairing-secret foobarbaz

# sign
keycard-generate-key
keycard-derive-key m/44'/60'/0'/0/0
keycard-sign 0000000000000000000000000000000000000000000000000000000000000000

# remove master key
keycard-remove-key

# unpair and check card status
keycard-unpair {{ session_pairing_index }}
keycard-select

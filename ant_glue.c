#include "ant_glue.h"
#include "ant_interface.h"
#include "ant_parameters.h"
#include "nrf_sdm.h"
#include <stdint.h>
#include <stddef.h>

uint8_t ANT_PLUS_NETWORK_KEY[] = {0xB9, 0xA5, 0x21, 0xFB, 0xBD, 0x72, 0xC3, 0x45};

void softdevice_fault_handler(uint32_t id, uint32_t pc, uint32_t info) {
    while(1);
}

uint32_t enable_softdevice() {
    uint32_t err;

    nrf_clock_lf_cfg_t clock_cfg = {
        .source        = NRF_CLOCK_LF_SRC_RC,
        .rc_ctiv       = 16,
        .rc_temp_ctiv  = 2,
        .accuracy      = 7
    };

    const char* license = "3831-521d-7df9-24d8-eff3-467b-225f-a00e";

    err = sd_softdevice_enable(&clock_cfg, softdevice_fault_handler, license);
    if (err != 0) return err;

    return err;
}

uint32_t setup_radar_channel() {
    uint32_t err;
    // 1. Set the Network Key on Network 0
    err = sd_ant_network_address_set(0, ANT_PLUS_NETWORK_KEY);
    if (err != 0) return err;

    // Channel 0, Master (0x10), Network 0, No extended assignment
    err = sd_ant_channel_assign(0, CHANNEL_TYPE_MASTER, 0, 0);
    if (err != 0) return err;

    // Device Number (random 16-bit), Device Type (40), Trans Type (1)
    err = sd_ant_channel_id_set(0, 10101, 40, 1);
    if (err != 0) return err;

    err = sd_ant_channel_radio_freq_set(0, 57);
    if (err != 0) return err;

    err = sd_ant_channel_period_set(0, 4084);
    if (err != 0) return err;

    return sd_ant_channel_open(0);
}

uint32_t send_radar_update(uint8_t * p_data) {
    // Channel 0, 8 bytes of data
    return sd_ant_broadcast_message_tx(0, 8, p_data);
}

void process_ant_events() {
    uint8_t ant_channel;
    uint8_t event_message_buffer[ANT_STANDARD_DATA_PAYLOAD_SIZE];

    while(sd_ant_event_get(&ant_channel, &event_message_buffer[0], NULL) == NRF_SUCCESS) {
        // Process or just clear the event
    }
}

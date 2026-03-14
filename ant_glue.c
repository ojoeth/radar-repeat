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

    static uint8_t ant_stack_buffer[512];
    ANT_ENABLE ant_enable_cfg = {
        .ucTotalNumberOfChannels = 5,
        .ucNumberOfEncryptedChannels = 0,
        .usNumberOfEvents = 48,
        .pucMemoryBlockStartLocation = ant_stack_buffer,
        .usMemoryBlockByteSize = sizeof(ant_stack_buffer)
    };


    err = sd_softdevice_enable(&clock_cfg, softdevice_fault_handler, license);
    if (err != 0) return err;

    err = sd_ant_enable(&ant_enable_cfg);
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
    err = sd_ant_channel_id_set(0, 10102, 40, 1);
    if (err != 0) return err;

    err = sd_ant_channel_radio_freq_set(0, 57);
    if (err != 0) return err;

    err = sd_ant_channel_period_set(0, 4084);
    if (err != 0) return err;

    return sd_ant_channel_open(0);
}

uint32_t setup_radar_receive_channel(uint32_t channel_num) {
    uint32_t err;

    // Clean up channel if it exists (ignore errors if not assigned)
    sd_ant_channel_close(channel_num);
    sd_ant_channel_unassign(channel_num);

    // Channel uses network 0 (already configured)
    err = sd_ant_channel_assign(channel_num, CHANNEL_TYPE_SLAVE, 0, 0);
    if (err != 0) return err+1000000;

    err = sd_ant_channel_id_set(channel_num, 0, 40, 0);
    if (err != 0) return err+2000000;

    err = sd_ant_channel_radio_freq_set(channel_num, 57);
    if (err != 0) return err+3000000;

    err = sd_ant_channel_period_set(channel_num, 4084);
    if (err != 0) return err+4000000;

    return sd_ant_channel_open(channel_num);
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

uint32_t get_ant_event(uint8_t *channel, uint8_t *event, uint8_t *data) {
    ANT_MESSAGE ant_msg;
    uint32_t err = sd_ant_event_get(channel, event, (uint8_t *)&ant_msg);

    if (err == 0) {
        // Payload[0] is channel, Payload[1-8] is the data
        for(int i = 0; i < 8; i++) {
            data[i] = ant_msg.ANT_MESSAGE_aucPayload[i];
        }
    }
    return err;
}

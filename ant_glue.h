#ifndef ANT_GLUE_H
#define ANT_GLUE_H

#include <stdint.h>

// Function prototypes
uint32_t setup_radar_channel(uint16_t device_num);
uint32_t enable_softdevice(void);
uint32_t send_radar_update(uint8_t *p_data);
uint32_t setup_radar_receive_channel(uint32_t channel_num);
uint32_t get_ant_event(uint8_t *channel, uint8_t *event, uint8_t *data);
uint32_t wait_for_ant_event(uint8_t *channel, uint8_t *event, uint8_t *data);
uint8_t get_channel_status(uint8_t channel);
void process_ant_events(void);
#endif

#ifndef ANT_GLUE_H
#define ANT_GLUE_H

#include <stdint.h>

// Function prototypes
uint32_t setup_radar_channel(void);
uint32_t enable_softdevice(void);
uint32_t send_radar_update(uint8_t *p_data);
void process_ant_events(void);
#endif

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <isa-l/erasure_code.h>

// #define K 3
// #define M 2
#define BLOCK_SIZE 65536


long getFileSize(FILE *file) {
    fseek(file, 0, SEEK_END);
    long size = ftell(file);
    rewind(file);
    return size;
}



int main(int argc,char *argv[]){
    if (argc!=4){
        fprintf(stderr,"Usage:: InputFile OuptutDir filename\n",argv[0]);
        return 1;
    }

    //initialize values argv[0] is the inputfile
    // K size and M Size (K+M) < 255

    // printf("Vals filename %s k %d and M is %d",argv[1],K,M);


    FILE *fin = fopen(argv[1],"rb");
    if (!fin){
        printf ("File not found or failed to open");
        return 1;
    }

    // read config file for setup values
    FILE *conffilein = fopen("config.txt","r");
    if (!conffilein){
        printf("Unable to open config file or file not present\n");
        return 1;
    }
    char key[128];
    char value[128];

    int K;
    int M;
    
    int values = 0;
    while (fscanf(conffilein,"%s %s",key,value)==2 && values<2){
        if (strcmp(key,"K")==0){
            K=atoi(value);
            // printf(" %s: %d \n",key,K);
        }else if(strcmp(key,"M")==0){
            M=atoi(value);
            // printf(" %s: %d \n",key,M);
        }
        values++;
    }

    //Get file size to set the block size.
    long fileSize = getFileSize(fin);
    
    //Instantiate the data pointers
    unsigned char *data[K];
    unsigned char *parity[M];

    //Memory assigned to heap linked to pointer data[i]
    for (int i=0;i<K;i++){
        data[i]=malloc(BLOCK_SIZE);
        if (!data[i]){
             perror("malloc data"); return 1; 
        }
    }

    //Memory assigned to heap linked to pointer parity[i]
    for (int i=0;i<M;i++){
        parity[i]=malloc(BLOCK_SIZE);
        if (!parity[i]){
            perror("mallic data");
            return 1;
        }
    }

    
    // Allocate encoding matrix and tables
    unsigned char *encode_matrix = malloc((K + M) * K);
    unsigned char *g_tbls = malloc(32 * K * M);
    if (!encode_matrix || !g_tbls) {
        perror("malloc encode_matrix or g_tbls");
        return 1;
    }

    // Generate matrix and initialize tables
    gf_gen_rs_matrix(encode_matrix, K + M, K);
    ec_init_tables(K, M, &encode_matrix[K * K], g_tbls);    

    size_t bytes_read;
    int block_num=0;
    while (1){
        int i;
        //Read K blocks into data buffers initialized.
        for (i=0;i<K;i++){
            bytes_read = fread(data[i],1,BLOCK_SIZE,fin);
            if (bytes_read < BLOCK_SIZE){
                // Zero pad the remaining data
                memset(data[i]+bytes_read,0,BLOCK_SIZE-bytes_read);
                break;
            }
        }

        if (i == 0 && bytes_read == 0) break; // EOF, no data read

        // Encode parity
        ec_encode_data(BLOCK_SIZE, K, M, g_tbls, data, parity);

        // Write data and parity blocks to files
        char filename[64];

        for (int j = 0; j < K; j++) {
            snprintf(filename, sizeof(filename), "%s/%s_data%d_block%d.bin",argv[2],argv[3], j, block_num);
            printf("%s_data%d_block%d.bin\n",argv[3], j, block_num);
            FILE *f = fopen(filename, "wb");
            if (!f) { perror("fopen data out"); return 1; }
            fwrite(data[j], 1, BLOCK_SIZE, f);
            fclose(f);
        }   

        for (int j = 0; j < M; j++) {
            snprintf(filename, sizeof(filename), "%s/%s_parity%d_block%d.bin",argv[2],argv[3], j, block_num);
            printf("%s_parity%d_block%d.bin\n",argv[3], j, block_num);
            FILE *f = fopen(filename, "wb");
            if (!f) { perror("fopen parity out"); return 1; }
            fwrite(parity[j], 1, BLOCK_SIZE, f);
            fclose(f);
        }

        //this marks the beginning of a new 3 blockSize-tuple 
        block_num++;

        //only part of the tuple got the data assigned and breaks loop
        // because even if part of data assigned it breaks out of the for Loop
        // this ensures that we have read and encoded to the last of the file.
        // ie for K=4. 7 4-64 byte tuples  arre assigned and 3 out of 4-64 byte tuple
        // got data so only thing to do here is to terminate the process
        if (i < K - 1) break;
    }

    //write meta data to file
    char metaFile[32];
    snprintf(metaFile, sizeof(metaFile), "metadata%d.txt", 1);

    FILE *mf = fopen(metaFile, "w");  // "w" since it's a text file
    if (!mf) {
        perror("Failed to open metadata file");
        return 1;
    }

    fprintf(mf, "blockSize: %ld\n", BLOCK_SIZE);
    fprintf(mf, "K: %d\n", K);
    fprintf(mf, "M: %d\n", M);  // Optional: include M if needed

    fclose(mf);

    //Clean up
    for (int i = 0; i < K; i++) free(data[i]);
    for (int i = 0; i < M; i++) free(parity[i]);
    free(encode_matrix);
    free(g_tbls);
    fclose(fin);

    return 0;
}
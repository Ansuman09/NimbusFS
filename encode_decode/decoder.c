#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <isa-l/erasure_code.h>

// #define K 3
// #define M 2
#define BLOCK_SIZE 65536 //64 Bytes

// Reads a file and assigns the BLOCK SIZE to memory for calculation
// IT is caught by the pointer buf that is passed by user.

int read_block(const char *filename, unsigned char *buf){
    FILE *f = fopen(filename,"rb");
    if (!f){
        //set the entire block size to zero buffer is of size BLOCK SIZE
        memset(buf,0,BLOCK_SIZE);
        return 0;
    }

    //assigns read bytes to buffer (space in memory)
    // and return the number of 1 bytes wrote
    size_t r = fread(buf,1,BLOCK_SIZE,f);
    fclose(f);
    //if number of bytes wrote is less than block size
    //that means space remains and can be zeroed out.
    if (r<BLOCK_SIZE){
        memset(buf+r,0,BLOCK_SIZE-r);
    }
    return 1;

}


int main(int argc, char *argv[]) {
    if (argc != 5) {
        fprintf(stderr, "Usage: %s outputfile inputfile_dir_and_name chunksize number_of_chunks: \n eg: %s recover /app/test/testfile\n", argv[0], argv[0]);
        return 1;
    }
    
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
            printf(" %s: %d \n",key,K);
        }else if(strcmp(key,"M")==0){
            M=atoi(value);
            printf(" %s: %d \n",key,M);
        }
        values++;
    }


    long chunksize = atol(argv[3]);
    int chunkNumber = atoi(argv[4]);
    //chars are used because the Intel ISA-L library is for 2^8 Galois feilds only
    unsigned char *data[K];
    unsigned char *coding[M];   
    unsigned char *recover[K+M];  //all blocks data+parity
    int present[K+M]; //for flagging wether block is available 0 for absent and 1 for present

    //allocate space for data , coding and recovery
    //data[0] == 03 d1 de 12 d12
    //data[1] == 7e ba 2f ac d12 etc
    
    for (int i=0;i<K;i++){
        data[i]=malloc(BLOCK_SIZE*sizeof(char));  // char is of size 1 Byte. 
        recover[i]=malloc(BLOCK_SIZE);   //malloc takes how many bytes of data . 
        if (!data[i] || !recover[i]){
            perror("malloc data");
            return 1;
        }
    }

    for (int i=0;i<M;i++){
        coding[i]=malloc(BLOCK_SIZE);
        recover[K+i] = malloc(BLOCK_SIZE);
        if (!coding[i] || !recover[K+i]){
            perror("malloc data");
            return 1;
        }
    }
    FILE *fout = fopen(argv[1], "r+b");
    if (!fout) {
        // File doesn't exist, create it
        fout = fopen(argv[1], "w+b");
        if (!fout) {
            perror("open output file");
            return 1;
        }
    }

    //seek the output pointer to the correct chunk 0 for success
    if (fseek(fout,chunkNumber*chunksize,SEEK_SET)!=0){
        perror("fseek failed");
        fclose(fout);
        return 1;
    }        



    //intitalize some matrices that we will use later
    unsigned char *encode_matrix = malloc ((K + M) * K);
    unsigned char *decode_matrix = malloc(K * K);
    unsigned char *invert_matrix = malloc(K * K);
    unsigned char *temp_matrix = malloc(K * K);
    unsigned char *g_tbls = malloc(32 * K * M);

    if (!encode_matrix || !decode_matrix || !invert_matrix || !temp_matrix || !g_tbls){
        perror("malloc matrices");
        return 1;
    }

    //generate the reed solomon matrix.
    gf_gen_rs_matrix(encode_matrix,K+M,K);
    

    //a counter that remembers which file is being read and how many array of
    // 64byte x K files.
    int block_num=0;

    while (1){
        //Read blocks for current stripe or data and mark those that are present
        int total_present = 0;

        //read the data block and pass it to recover array
        // recover[1] = data_block 1
        // recover[2] = data_block_2
        // ...
        // increased present block counter
        for (int i=0;i<K;i++){  
            char filename[129];
            snprintf(filename, sizeof(filename), "%s_data%d_block%d.bin", argv[2],i, block_num);
            present[i] = read_block(filename, recover[i]);
            total_present += present[i];
        }

        for (int i = 0; i < M; i++) {
            char filename[128];
            snprintf(filename, sizeof(filename), "%s_parity%d_block%d.bin",argv[2] ,i, block_num);
            present[K + i] = read_block(filename, recover[K + i]);
            total_present += present[K + i];
        }

        if (total_present ==0 ) break; // no more blocks

        if (total_present < K){
            fprintf(stderr,"ERROR: Not enought blocks to recover the file");
        }

        // Initialize the decode indices or indexes that can be used to decode.
        // decode index however holds the index of the files that are available to decode

        int decode_idx[K];
        int di =0;
        for (int i=0;i<K+M && di<K;i++){
            if (present[i]){
                decode_idx[di++]=i;
            }
        }

        //copy over elements generated for  encode matruc to decore matrix
        for (int i =0;i<K;i++){
            memcpy(&decode_matrix[i*K],&encode_matrix[decode_idx[i]*K],K);
        }

        //now invert this matrix to get the decode matrix 
        // we store the data in invert matrix now.
        if (gf_invert_matrix(decode_matrix, invert_matrix, K) < 0) {
            fprintf(stderr, "Failed to invert decode matrix\n");
            return 1;
        }

        //assign the same block size to heap and point it with temp
        unsigned char *tmp = malloc(BLOCK_SIZE);
        unsigned char gftbl[32];
        
        if (!tmp){
            perror("malloc tmp");
            return 1;
        }
        
        //start preparing the final data. initialize every data block to zeroes

        for (int i=0;i<K;i++){
            memset(data[i],0,BLOCK_SIZE);

            // Here the matrix is an entire array, so 
            // if size is 3x3 matrix then elemeents 
            // are in (0,11),(1,12),(2,13),(3,21),(4,22),(5,23),(6,31),(7,32),(8,33).
        
            for (int j=0; j<K ; j++){
                unsigned char coef = invert_matrix[i*K+j];
                if (coef!=0){
                    ec_init_tables(1,1,&coef,gftbl);

                    //multiply coefficeint to recover block and store in tmp.
                    // now this mult must follow Galois Field (2^8) rules while getting the product
                    // which we will later use to mult
                    // this essentialy does invert mat  row val [a1] dot product [recover[i]th col]
                    // This is how it visually looks like::
                    // a1*e2 a1*34 a1*0d a1*6f a1*ba ...  where vector row [e2,34,0d,gf,ba...] are part of the active
                    // data available for decoding.
                   gf_vect_mul(BLOCK_SIZE, gftbl, recover[decode_idx[j]], tmp);


                   // these values are now stored in the data[i] lets say for the first iteration data[1]
                   // then data[1]=[a1*e2 a1*34 a1*0d a1*6f a1*ba  ..]
                   // then for a2 and recover[d2] = [0b 1b 5b 6d b3 ea ...]
                   // data[1]=[a1*e2 XOR a2*01b ; a1*34 XOR a2*1b ; a1*0d XOR a2*5b ; a1*6f XOR d2*6d ...]
                   for (int b =0;b< BLOCK_SIZE;b++){
                    data[i][b] ^= tmp[b]; // follow XOR MULT
                   }
                }
            }
            // so we get our resulting data one for each iteration of this for-loop.
                   
        }

        free(tmp);
        //write data per byte from data to out
        // essentially we are copying over data[i] to out then
        // data[i+1] to out ... and so on.
        for (int i = 0; i < K; i++) {
            fwrite(data[i], 1, BLOCK_SIZE, fout);
        }

        block_num++;
    
    }


    //free up allocated data.
    printf("Decoding complete. Total stripes: %d\n", block_num);

    for (int i = 0; i < K; i++) {
        free(data[i]);
        free(recover[i]);
    }
    for (int i = 0; i < M; i++) {
        free(coding[i]);
        free(recover[K + i]);
    }

    free(encode_matrix);
    free(decode_matrix);
    free(invert_matrix);
    free(temp_matrix);
    free(g_tbls);

    fclose(fout);

    return 0;

    
 }
